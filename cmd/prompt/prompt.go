package prompt

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/briandowns/spinner"
	"github.com/cschleiden/promptmd"
	"github.com/github/gh-models/internal/azuremodels"
	"github.com/github/gh-models/internal/sse"
	"github.com/github/gh-models/pkg/command"
	"github.com/github/gh-models/pkg/util"
	"github.com/spf13/cobra"
)

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

// NewPromptCommand returns a new gh command for prompting a model using a prompt from a file.
func NewPromptCommand(cfg *command.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prompt [prompt.md] [key=value]*",
		Short: "Run inference with the specified model",
		Long: heredoc.Docf(`
			Prompts the specified model with the given prompt. Replace any {{placeholders}} in the prompts with "key=value" arguments.
		`),
		Example: "gh models prompt my-prompt.prompt.md",
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdHandler := newPromptCommandHandler(cmd, cfg, args)
			if cmdHandler == nil {
				return nil
			}

			models, err := cmdHandler.loadModels()
			if err != nil {
				return err
			}

			var promptPath string

			if len(args) > 0 {
				promptPath = args[0]
			}

			promptContent, err := os.ReadFile(promptPath)
			if err != nil {
				return err
			}

			prompt, err := promptmd.Parse(string(promptContent))
			if err != nil {
				return err
			}

			var modelName string
			if prompt.Metadata != nil {
				m, ok := prompt.Metadata["model"]
				if ok {
					modelName = m.(string)
				}
			}

			if modelName == "" {
				modelName, err = cmdHandler.promptForModelName(models)
				if err != nil {
					return err
				}
			}

			// Try to parse any key=value pairs from args
			vars := make(promptmd.Vars)
			for _, arg := range args[1:] {
				parts := strings.SplitN(arg, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid argument: %s, expected key=value", arg)
				}

				vars[parts[0]] = parts[1]
			}

			conversation := Conversation{}

			pp, err := prompt.Prepare()
			if err != nil {
				return fmt.Errorf("failed to prepare prompt: %w", err)
			}

			msgs, err := pp(vars)
			if err != nil {
				return fmt.Errorf("failed to prepare prompt messages: %w", err)
			}

			for _, message := range msgs {
				if message.Role == promptmd.RoleSystem {
					conversation.systemPrompt = message.Message
				} else {
					conversation.AddMessage(azuremodels.ChatMessageRole(message.Role), message.Message)
				}
			}

			req := azuremodels.ChatCompletionOptions{
				Messages: conversation.GetMessages(),
				Model:    modelName,
			}

			sp := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(cmdHandler.cfg.ErrOut))
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

			cmdHandler.writeToOut("\n")
			_, err = messageBuilder.WriteString("\n")
			if err != nil {
				return err
			}

			conversation.AddMessage(azuremodels.ChatMessageRoleAssistant, messageBuilder.String())

			return nil
		},
	}

	return cmd
}

type runCommandHandler struct {
	ctx    context.Context
	cfg    *command.Config
	client azuremodels.Client
	args   []string
}

func newPromptCommandHandler(cmd *cobra.Command, cfg *command.Config, args []string) *runCommandHandler {
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

func (h *runCommandHandler) promptForModelName(models []*azuremodels.ModelSummary) (string, error) {
	modelName := ""

	// Need to prompt for a model
	prompt := &survey.Select{
		Message: "Select a model:",
		Options: []string{},
	}

	for _, model := range models {
		if !model.IsChatModel() {
			continue
		}
		prompt.Options = append(prompt.Options, model.FriendlyName)
	}

	err := survey.AskOne(prompt, &modelName, survey.WithPageSize(10))
	if err != nil {
		return "", err
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
