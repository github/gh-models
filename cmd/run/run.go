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
	"github.com/briandowns/spinner"
	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/cli/go-gh/v2/pkg/term"
	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/internal/sse"
	"github.com/github/gh-models/internal/ux"
	"github.com/github/gh-models/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

// NewRunCommand returns a new gh command for running a model.
func NewRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [model] [prompt]",
		Short: "Run inference with the specified model",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdHandler := newRunCommandHandler(cmd, args)
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

			if len(args) > 1 {
				initialPrompt = strings.Join(args[1:], " ")
				singleShot = true
			}

			if isPipe(os.Stdin) {
				promptFromPipe, _ := io.ReadAll(os.Stdin)
				if len(promptFromPipe) > 0 {
					initialPrompt = initialPrompt + "\n" + string(promptFromPipe)
					singleShot = true
				}
			}

			systemPrompt, err := cmd.Flags().GetString("system-prompt")
			if err != nil {
				return err
			}

			conversation := Conversation{
				systemPrompt: systemPrompt,
			}

			mp := ModelParameters{}
			err = mp.PopulateFromFlags(cmd.Flags())
			if err != nil {
				return err
			}

			for {
				prompt := ""
				if initialPrompt != "" {
					prompt = initialPrompt
					initialPrompt = ""
				}

				if prompt == "" {
					fmt.Printf(">>> ")
					reader := bufio.NewReader(os.Stdin)
					prompt, err = reader.ReadString('\n')
					if err != nil {
						return err
					}
				}

				prompt = strings.TrimSpace(prompt)

				if prompt == "" {
					continue
				}

				if strings.HasPrefix(prompt, "/") {
					if prompt == "/bye" || prompt == "/exit" || prompt == "/quit" {
						break
					}

					if prompt == "/parameters" {
						cmdHandler.handleParametersPrompt(conversation, mp)
						continue
					}

					if prompt == "/reset" || prompt == "/clear" {
						cmdHandler.handleResetPrompt(conversation)
						continue
					}

					if strings.HasPrefix(prompt, "/set ") {
						cmdHandler.handleSetPrompt(prompt, mp)
						continue
					}

					if strings.HasPrefix(prompt, "/system-prompt ") {
						conversation = cmdHandler.handleSystemPrompt(prompt, conversation)
						continue
					}

					if prompt == "/help" {
						cmdHandler.handleHelpPrompt()
						continue
					}

					cmdHandler.handleUnrecognizedPrompt(prompt)
					continue
				}

				conversation.AddMessage(azuremodels.ChatMessageRoleUser, prompt)

				req := azuremodels.ChatCompletionOptions{
					Messages: conversation.GetMessages(),
					Model:    modelName,
				}

				mp.UpdateRequest(&req)

				sp := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(cmdHandler.errOut))
				sp.Start()
				defer sp.Stop()

				reader, err := cmdHandler.getChatCompletionStreamReader(req)
				if err != nil {
					return err
				}
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

				util.WriteToOut(cmdHandler.out, "\n")
				_, err = messageBuilder.WriteString("\n")
				if err != nil {
					return err
				}

				conversation.AddMessage(azuremodels.ChatMessageRoleAssistant, messageBuilder.String())

				if singleShot {
					break
				}
			}

			return nil
		},
	}

	cmd.Flags().String("max-tokens", "", "Limit the maximum tokens for the model response.")
	cmd.Flags().String("temperature", "", "Controls randomness in the response, use lower to be more deterministic.")
	cmd.Flags().String("top-p", "", "Controls text diversity by selecting the most probable words until a set probability is reached.")
	cmd.Flags().String("system-prompt", "", "Prompt the system.")

	return cmd
}

type runCommandHandler struct {
	ctx      context.Context
	terminal term.Term
	out      io.Writer
	errOut   io.Writer
	client   *azuremodels.Client
	args     []string
}

func newRunCommandHandler(cmd *cobra.Command, args []string) *runCommandHandler {
	terminal := term.FromEnv()
	out := terminal.Out()
	token, _ := auth.TokenForHost("github.com")
	if token == "" {
		util.WriteToOut(out, "No GitHub token found. Please run 'gh auth login' to authenticate.\n")
		return nil
	}
	return &runCommandHandler{
		ctx:      cmd.Context(),
		terminal: terminal,
		out:      out,
		args:     args,
		errOut:   terminal.ErrOut(),
		client:   azuremodels.NewClient(token),
	}
}

func (h *runCommandHandler) loadModels() ([]*azuremodels.ModelSummary, error) {
	models, err := h.client.ListModels(h.ctx)
	if err != nil {
		return nil, err
	}

	ux.SortModels(models)
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
			if !ux.IsChatModel(model) {
				continue
			}
			prompt.Options = append(prompt.Options, model.FriendlyName)
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
			modelName = model.Name
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
	util.WriteToOut(h.out, "Current parameters:\n")
	names := []string{"max-tokens", "temperature", "top-p"}
	for _, name := range names {
		util.WriteToOut(h.out, fmt.Sprintf("  %s: %s\n", name, mp.FormatParameter(name)))
	}
	util.WriteToOut(h.out, "\n")
	util.WriteToOut(h.out, "System Prompt:\n")
	if conversation.systemPrompt != "" {
		util.WriteToOut(h.out, "  "+conversation.systemPrompt+"\n")
	} else {
		util.WriteToOut(h.out, "  <not set>\n")
	}
}

func (h *runCommandHandler) handleResetPrompt(conversation Conversation) {
	conversation.Reset()
	util.WriteToOut(h.out, "Reset chat history\n")
}

func (h *runCommandHandler) handleSetPrompt(prompt string, mp ModelParameters) {
	parts := strings.Split(prompt, " ")
	if len(parts) == 3 {
		name := parts[1]
		value := parts[2]

		err := mp.SetParameterByName(name, value)
		if err != nil {
			util.WriteToOut(h.out, err.Error()+"\n")
			return
		}

		util.WriteToOut(h.out, "Set "+name+" to "+value+"\n")
	} else {
		util.WriteToOut(h.out, "Invalid /set syntax. Usage: /set <name> <value>\n")
	}
}

func (h *runCommandHandler) handleSystemPrompt(prompt string, conversation Conversation) Conversation {
	conversation.systemPrompt = strings.Trim(strings.TrimPrefix(prompt, "/system-prompt "), "\"")
	util.WriteToOut(h.out, "Updated system prompt\n")
	return conversation
}

func (h *runCommandHandler) handleHelpPrompt() {
	util.WriteToOut(h.out, "Commands:\n")
	util.WriteToOut(h.out, "  /bye, /exit, /quit - Exit the chat\n")
	util.WriteToOut(h.out, "  /parameters - Show current model parameters\n")
	util.WriteToOut(h.out, "  /reset, /clear - Reset chat context\n")
	util.WriteToOut(h.out, "  /set <name> <value> - Set a model parameter\n")
	util.WriteToOut(h.out, "  /system-prompt <prompt> - Set the system prompt\n")
	util.WriteToOut(h.out, "  /help - Show this help message\n")
}

func (h *runCommandHandler) handleUnrecognizedPrompt(prompt string) {
	util.WriteToOut(h.out, "Unknown command '"+prompt+"'. See /help for supported commands.\n")
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
		util.WriteToOut(h.out, *content)
	} else if choice.Message != nil && choice.Message.Content != nil {
		content := choice.Message.Content
		_, err := messageBuilder.WriteString(*content)
		if err != nil {
			return err
		}
		util.WriteToOut(h.out, *content)
	}

	// Introduce a small delay in between response tokens to better simulate a conversation
	if h.terminal.IsTerminalOutput() {
		time.Sleep(10 * time.Millisecond)
	}

	return nil
}
