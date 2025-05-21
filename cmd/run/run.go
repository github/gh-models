// Package run provides a gh command to run a GitHub model.
package run

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/briandowns/spinner"
	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/internal/sse"
	"github.com/github/gh-models/pkg/command"
	"github.com/github/gh-models/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

// ModelParameters represents the parameters that can be set for a model run.
type ModelParameters struct {
	maxTokens   *int
	temperature *float64
	topP        *float64
}

// FormatParameter returns a string representation of the parameter value.
func (mp *ModelParameters) FormatParameter(name string) string {
	switch name {
	case "max-tokens":
		if mp.maxTokens != nil {
			return strconv.Itoa(*mp.maxTokens)
		}

	case "temperature":
		if mp.temperature != nil {
			return fmt.Sprintf("%f", *mp.temperature)
		}

	case "top-p":
		if mp.topP != nil {
			return fmt.Sprintf("%f", *mp.topP)
		}
	}

	return "<not set>"
}

// PopulateFromFlags populates the model parameters from the given flags.
func (mp *ModelParameters) PopulateFromFlags(flags *pflag.FlagSet) error {
	maxTokensString, err := flags.GetString("max-tokens")
	if err != nil {
		return err
	}
	if maxTokensString != "" {
		maxTokens, err := strconv.Atoi(maxTokensString)
		if err != nil {
			return err
		}
		mp.maxTokens = util.Ptr(maxTokens)
	}

	temperatureString, err := flags.GetString("temperature")
	if err != nil {
		return err
	}
	if temperatureString != "" {
		temperature, err := strconv.ParseFloat(temperatureString, 64)
		if err != nil {
			return err
		}
		mp.temperature = util.Ptr(temperature)
	}

	topPString, err := flags.GetString("top-p")
	if err != nil {
		return err
	}
	if topPString != "" {
		topP, err := strconv.ParseFloat(topPString, 64)
		if err != nil {
			return err
		}
		mp.topP = util.Ptr(topP)
	}

	return nil
}

// SetParameterByName sets the parameter with the given name to the given value.
func (mp *ModelParameters) SetParameterByName(name, value string) error {
	switch name {
	case "max-tokens":
		maxTokens, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		mp.maxTokens = util.Ptr(maxTokens)

	case "temperature":
		temperature, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		mp.temperature = util.Ptr(temperature)

	case "top-p":
		topP, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		mp.topP = util.Ptr(topP)

	default:
		return errors.New("unknown parameter '" + name + "'. Supported parameters: max-tokens, temperature, top-p")
	}

	return nil
}

// UpdateRequest updates the given request with the model parameters.
func (mp *ModelParameters) UpdateRequest(req *azuremodels.ChatCompletionOptions) {
	req.MaxTokens = mp.maxTokens
	req.Temperature = mp.temperature
	req.TopP = mp.topP
}

// Conversation represents a conversation between the user and the model.
type Conversation struct {
	messages     []azuremodels.ChatMessage
	systemPrompt string
}

// AddMessage adds a message to the conversation.
func (c *Conversation) AddMessage(role azuremodels.ChatMessageRole, content string) {
	c.messages = append(c.messages, azuremodels.ChatMessage{
		Content: util.Ptr(content),
		Role:    role,
	})
}

// GetMessages returns the messages in the conversation.
func (c *Conversation) GetMessages() []azuremodels.ChatMessage {
	length := len(c.messages)
	if c.systemPrompt != "" {
		length++
	}

	messages := make([]azuremodels.ChatMessage, length)
	startIndex := 0

	if c.systemPrompt != "" {
		messages[0] = azuremodels.ChatMessage{
			Content: util.Ptr(c.systemPrompt),
			Role:    azuremodels.ChatMessageRoleSystem,
		}
		startIndex++
	}

	for i, message := range c.messages {
		messages[startIndex+i] = message
	}

	return messages
}

// Reset removes messages from the conversation.
func (c *Conversation) Reset() {
	c.messages = nil
}

func isPipe(r io.Reader) bool {
	if f, ok := r.(*os.File); ok {
		stat, err := f.Stat()
		if err != nil {
			return false
		}
		if stat.Mode()&os.ModeNamedPipe != 0 {
			return true
		}
	}
	return false
}

// promptFile mirrors the format of .prompt.yml
type promptFile struct {
	Name            string `yaml:"name"`
	Description     string `yaml:"description"`
	Model           string `yaml:"model"`
	ModelParameters struct {
		MaxTokens   *int     `yaml:"maxTokens"`
		Temperature *float64 `yaml:"temperature"`
		TopP        *float64 `yaml:"topP"`
	} `yaml:"modelParameters"`
	Messages []struct {
		Role    string `yaml:"role"`
		Content string `yaml:"content"`
	} `yaml:"messages"`
}

// NewRunCommand returns a new gh command for running a model.
func NewRunCommand(cfg *command.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [model] [prompt]",
		Short: "Run inference with the specified model",
		Long: heredoc.Docf(`
			Prompts the specified model with the given prompt.

			Use %[1]sgh models run%[1]s to run in interactive mode. It will provide a list of the current
			models and allow you to select the one you want to run an inference with. After you select the model
			you will be able to enter the prompt you want to run via the selected model.

			If you know which model you want to run inference with, you can run the request in a single command
			as %[1]sgh models run [model] [prompt]%[1]s

			The return value will be the response to your prompt from the selected model.
		`, "`"),
		Example: "gh models run openai/gpt-4o-mini \"how many types of hyena are there?\"",
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath, _ := cmd.Flags().GetString("file")
			var pf *promptFile
			if filePath != "" {
				b, err := os.ReadFile(filePath)
				if err != nil {
					return err
				}
				p := promptFile{}
				if err := yaml.Unmarshal(b, &p); err != nil {
					return err
				}
				pf = &p
				// Inject model name as the first positional arg if user didn't supply one
				if pf.Model != "" && len(args) == 0 {
					args = append([]string{pf.Model}, args...)
				}
			}

			cmdHandler := newRunCommandHandler(cmd, cfg, args)
			if cmdHandler == nil {
				return nil
			}

			models, err := cmdHandler.loadModels()
			if err != nil {
				return err
			}

			modelName, err := cmdHandler.getModelNameFromArgs(models)
			if err != nil {
				return err
			}

			initialPrompt := ""
			singleShot := false
			pipedContent := ""

			if len(args) > 1 {
				initialPrompt = strings.Join(args[1:], " ")
				singleShot = true
			}

			if isPipe(os.Stdin) {
				promptFromPipe, _ := io.ReadAll(os.Stdin)
				if len(promptFromPipe) > 0 {
					singleShot = true
					pipedContent = strings.TrimSpace(string(promptFromPipe))
					if initialPrompt != "" {
						initialPrompt = initialPrompt + "\n" + pipedContent
					} else {
						initialPrompt = pipedContent
					}
				}
			}

			systemPrompt, err := cmd.Flags().GetString("system-prompt")
			if err != nil {
				return err
			}

			conversation := Conversation{
				systemPrompt: systemPrompt,
			}

			// If there is no prompt file, add the initialPrompt to the conversation.
			// If a prompt file is passed, load the messages from the file, templating {{input}}
			// using the initialPrompt.
			if pf == nil {
				conversation.AddMessage(azuremodels.ChatMessageRoleUser, initialPrompt)
			} else {
				for _, m := range pf.Messages {
					content := m.Content
					switch strings.ToLower(m.Role) {
					case "system":
						conversation.systemPrompt = content
					case "user":
						content = strings.ReplaceAll(content, "{{input}}", initialPrompt)
						conversation.AddMessage(azuremodels.ChatMessageRoleUser, content)
					case "assistant":
						conversation.AddMessage(azuremodels.ChatMessageRoleAssistant, content)
					}
				}
			}

			mp := ModelParameters{}
			err = mp.PopulateFromFlags(cmd.Flags())
			if err != nil {
				return err
			}

			if pf != nil {
				mp.maxTokens = pf.ModelParameters.MaxTokens
				mp.temperature = pf.ModelParameters.Temperature
				mp.topP = pf.ModelParameters.TopP
			}

			for {
				if !singleShot {
					conversation, err = cmdHandler.ChatWithUser(conversation, mp)
					if errors.Is(err, ExitChatError) {
						break
					} else if err != nil {
						return err
					}
				}

				req := azuremodels.ChatCompletionOptions{
					Messages: conversation.GetMessages(),
					Model:    modelName,
				}

				mp.UpdateRequest(&req)

				sp := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(cmdHandler.cfg.ErrOut))
				sp.Start()
				//nolint:gocritic,revive // TODO
				defer sp.Stop()

				reader, err := cmdHandler.getChatCompletionStreamReader(req)
				if err != nil {
					return err
				}
				//nolint:gocritic,revive // TODO
				defer reader.Close()

				messageBuilder := strings.Builder{}

				for {
					completion, err := reader.Read()
					if err != nil {
						if errors.Is(err, io.EOF) {
							break
						}
						return err
					}

					sp.Stop()

					for _, choice := range completion.Choices {
						err = cmdHandler.handleCompletionChoice(choice, messageBuilder)
						if err != nil {
							return err
						}
					}
				}

				cmdHandler.writeToOut("\n")
				_, err = messageBuilder.WriteString("\n")
				if err != nil {
					return err
				}

				conversation.AddMessage(azuremodels.ChatMessageRoleAssistant, messageBuilder.String())

				if singleShot || pf != nil {
					break
				}
			}

			return nil
		},
	}

	cmd.Flags().String("file", "", "Path to a .prompt.yml file.")
	cmd.Flags().String("max-tokens", "", "Limit the maximum tokens for the model response.")
	cmd.Flags().String("temperature", "", "Controls randomness in the response, use lower to be more deterministic.")
	cmd.Flags().String("top-p", "", "Controls text diversity by selecting the most probable words until a set probability is reached.")
	cmd.Flags().String("system-prompt", "", "Prompt the system.")

	return cmd
}

type runCommandHandler struct {
	ctx    context.Context
	cfg    *command.Config
	client azuremodels.Client
	args   []string
}

func newRunCommandHandler(cmd *cobra.Command, cfg *command.Config, args []string) *runCommandHandler {
	return &runCommandHandler{ctx: cmd.Context(), cfg: cfg, client: cfg.Client, args: args}
}

func (h *runCommandHandler) loadModels() ([]*azuremodels.ModelSummary, error) {
	models, err := h.client.ListModels(h.ctx)
	if err != nil {
		return nil, err
	}

	azuremodels.SortModels(models)
	return models, nil
}

func (h *runCommandHandler) getModelNameFromArgs(models []*azuremodels.ModelSummary) (string, error) {
	modelName := ""

	switch {
	case len(h.args) == 0:
		// Need to prompt for a model
		prompt := &survey.Select{
			Message: "Select a model:",
			Options: []string{},
		}

		for _, model := range models {
			if !model.IsChatModel() {
				continue
			}
			prompt.Options = append(prompt.Options, azuremodels.FormatIdentifier(model.Publisher, model.Name))
		}

		err := survey.AskOne(prompt, &modelName, survey.WithPageSize(10))
		if err != nil {
			return "", err
		}

	case len(h.args) >= 1:
		modelName = h.args[0]
	}

	return validateModelName(modelName, models)
}

func validateModelName(modelName string, models []*azuremodels.ModelSummary) (string, error) {
	noMatchErrorMessage := "The specified model name is not found. Run 'gh models list' to see available models or 'gh models run' to select interactively."

	if modelName == "" {
		return "", errors.New(noMatchErrorMessage)
	}

	foundMatch := false
	for _, model := range models {
		if model.HasName(modelName) {
			foundMatch = true
			break
		}
	}

	if !foundMatch {
		return "", errors.New(noMatchErrorMessage)
	}

	return modelName, nil
}

func (h *runCommandHandler) getChatCompletionStreamReader(req azuremodels.ChatCompletionOptions) (sse.Reader[azuremodels.ChatCompletion], error) {
	resp, err := h.client.GetChatCompletionStream(h.ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Reader, nil
}

func (h *runCommandHandler) handleParametersPrompt(conversation Conversation, mp ModelParameters) {
	h.writeToOut("Current parameters:\n")
	names := []string{"max-tokens", "temperature", "top-p"}
	for _, name := range names {
		h.writeToOut(fmt.Sprintf("  %s: %s\n", name, mp.FormatParameter(name)))
	}
	h.writeToOut("\n")
	h.writeToOut("System Prompt:\n")
	if conversation.systemPrompt != "" {
		h.writeToOut("  " + conversation.systemPrompt + "\n")
	} else {
		h.writeToOut("  <not set>\n")
	}
}

func (h *runCommandHandler) handleResetPrompt(conversation Conversation) {
	conversation.Reset()
	h.writeToOut("Reset chat history\n")
}

func (h *runCommandHandler) handleSetPrompt(prompt string, mp ModelParameters) {
	parts := strings.Split(prompt, " ")
	if len(parts) == 3 {
		name := parts[1]
		value := parts[2]

		err := mp.SetParameterByName(name, value)
		if err != nil {
			h.writeToOut(err.Error() + "\n")
			return
		}

		h.writeToOut("Set " + name + " to " + value + "\n")
	} else {
		h.writeToOut("Invalid /set syntax. Usage: /set <name> <value>\n")
	}
}

func (h *runCommandHandler) handleSystemPrompt(prompt string, conversation Conversation) Conversation {
	conversation.systemPrompt = strings.Trim(strings.TrimPrefix(prompt, "/system-prompt "), "\"")
	h.writeToOut("Updated system prompt\n")
	return conversation
}

func (h *runCommandHandler) handleHelpPrompt() {
	h.writeToOut("Commands:\n")
	h.writeToOut("  /bye, /exit, /quit - Exit the chat\n")
	h.writeToOut("  /parameters - Show current model parameters\n")
	h.writeToOut("  /reset, /clear - Reset chat context\n")
	h.writeToOut("  /set <name> <value> - Set a model parameter\n")
	h.writeToOut("  /system-prompt <prompt> - Set the system prompt\n")
	h.writeToOut("  /help - Show this help message\n")
}

func (h *runCommandHandler) handleUnrecognizedPrompt(prompt string) {
	h.writeToOut("Unknown command '" + prompt + "'. See /help for supported commands.\n")
}

func (h *runCommandHandler) handleCompletionChoice(choice azuremodels.ChatChoice, messageBuilder strings.Builder) error {
	// Streamed responses from the OpenAI API have their data in `.Delta`, while
	// non-streamed responses use `.Message`, so let's support both
	if choice.Delta != nil && choice.Delta.Content != nil {
		content := choice.Delta.Content
		_, err := messageBuilder.WriteString(*content)
		if err != nil {
			return err
		}
		h.writeToOut(*content)
	} else if choice.Message != nil && choice.Message.Content != nil {
		content := choice.Message.Content
		_, err := messageBuilder.WriteString(*content)
		if err != nil {
			return err
		}
		h.writeToOut(*content)
	}

	// Introduce a small delay in between response tokens to better simulate a conversation
	if h.cfg.IsTerminalOutput {
		time.Sleep(10 * time.Millisecond)
	}

	return nil
}

func (h *runCommandHandler) writeToOut(message string) {
	h.cfg.WriteToOut(message)
}

var ExitChatError = errors.New("exiting chat")

func (h *runCommandHandler) ChatWithUser(conversation Conversation, mp ModelParameters) (Conversation, error) {
	fmt.Printf(">>> ")
	reader := bufio.NewReader(os.Stdin)

	prompt, err := reader.ReadString('\n')
	if err != nil {
		return conversation, err
	}

	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return conversation, nil
	}

	if strings.HasPrefix(prompt, "/") {
		if prompt == "/bye" || prompt == "/exit" || prompt == "/quit" {
			return conversation, ExitChatError
		}

		if prompt == "/parameters" {
			h.handleParametersPrompt(conversation, mp)
			return conversation, nil
		}

		if prompt == "/reset" || prompt == "/clear" {
			h.handleResetPrompt(conversation)
			return conversation, nil
		}

		if strings.HasPrefix(prompt, "/set ") {
			h.handleSetPrompt(prompt, mp)
			return conversation, nil
		}

		if strings.HasPrefix(prompt, "/system-prompt ") {
			conversation = h.handleSystemPrompt(prompt, conversation)
			return conversation, nil
		}

		if prompt == "/help" {
			h.handleHelpPrompt()
			return conversation, nil
		}

		h.handleUnrecognizedPrompt(prompt)
		return conversation, nil
	}

	conversation.AddMessage(azuremodels.ChatMessageRoleUser, prompt)
	return conversation, nil
}
