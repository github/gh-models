// run provides a gh command to run a GitHub model.
package run

import (
	"bufio"
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
	"github.com/github/gh-models/internal/azure_models"
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
		mp.maxTokens = azure_models.Ptr(maxTokens)
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
		mp.temperature = azure_models.Ptr(temperature)
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
		mp.topP = azure_models.Ptr(topP)
	}

	return nil
}

// SetParameterByName sets the parameter with the given name to the given value.
func (mp *ModelParameters) SetParameterByName(name string, value string) error {
	switch name {
	case "max-tokens":
		maxTokens, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		mp.maxTokens = azure_models.Ptr(maxTokens)

	case "temperature":
		temperature, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		mp.temperature = azure_models.Ptr(temperature)

	case "top-p":
		topP, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		mp.topP = azure_models.Ptr(topP)

	default:
		return errors.New("unknown parameter '" + name + "'. Supported parameters: max-tokens, temperature, top-p")
	}

	return nil
}

// UpdateRequest updates the given request with the model parameters.
func (mp *ModelParameters) UpdateRequest(req *azure_models.ChatCompletionOptions) {
	req.MaxTokens = mp.maxTokens
	req.Temperature = mp.temperature
	req.TopP = mp.topP
}

// Conversation represents a conversation between the user and the model.
type Conversation struct {
	messages     []azure_models.ChatMessage
	systemPrompt string
}

// AddMessage adds a message to the conversation.
func (c *Conversation) AddMessage(role azure_models.ChatMessageRole, content string) {
	c.messages = append(c.messages, azure_models.ChatMessage{
		Content: azure_models.Ptr(content),
		Role:    role,
	})
}

// GetMessages returns the messages in the conversation.
func (c *Conversation) GetMessages() []azure_models.ChatMessage {
	length := len(c.messages)
	if c.systemPrompt != "" {
		length++
	}

	messages := make([]azure_models.ChatMessage, length)
	startIndex := 0

	if c.systemPrompt != "" {
		messages[0] = azure_models.ChatMessage{
			Content: azure_models.Ptr(c.systemPrompt),
			Role:    azure_models.ChatMessageRoleSystem,
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
			terminal := term.FromEnv()
			out := terminal.Out()
			errOut := terminal.ErrOut()

			token, _ := auth.TokenForHost("github.com")
			if token == "" {
				util.WriteToOut(out, "No GitHub token found. Please run 'gh auth login' to authenticate.\n")
				return nil
			}

			client := azure_models.NewClient(token)

			models, err := client.ListModels()
			if err != nil {
				return err
			}

			ux.SortModels(models)

			modelName := ""
			switch {
			case len(args) == 0:
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

				err = survey.AskOne(prompt, &modelName, survey.WithPageSize(10))
				if err != nil {
					return err
				}

			case len(args) >= 1:
				modelName = args[0]
			}

			noMatchErrorMessage := "The specified model name is not found. Run 'gh models list' to see available models or 'gh models run' to select interactively."

			if modelName == "" {
				return errors.New(noMatchErrorMessage)
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
				return errors.New(noMatchErrorMessage)
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
						util.WriteToOut(out, "Current parameters:\n")
						names := []string{"max-tokens", "temperature", "top-p"}
						for _, name := range names {
							util.WriteToOut(out, fmt.Sprintf("  %s: %s\n", name, mp.FormatParameter(name)))
						}
						util.WriteToOut(out, "\n")
						util.WriteToOut(out, "System Prompt:\n")
						if conversation.systemPrompt != "" {
							util.WriteToOut(out, "  "+conversation.systemPrompt+"\n")
						} else {
							util.WriteToOut(out, "  <not set>\n")
						}
						continue
					}

					if prompt == "/reset" || prompt == "/clear" {
						conversation.Reset()
						util.WriteToOut(out, "Reset chat history\n")
						continue
					}

					if strings.HasPrefix(prompt, "/set ") {
						parts := strings.Split(prompt, " ")
						if len(parts) == 3 {
							name := parts[1]
							value := parts[2]

							err := mp.SetParameterByName(name, value)
							if err != nil {
								util.WriteToOut(out, err.Error()+"\n")
								continue
							}

							util.WriteToOut(out, "Set "+name+" to "+value+"\n")
						} else {
							util.WriteToOut(out, "Invalid /set syntax. Usage: /set <name> <value>\n")
						}
						continue
					}

					if strings.HasPrefix(prompt, "/system-prompt ") {
						conversation.systemPrompt = strings.Trim(strings.TrimPrefix(prompt, "/system-prompt "), "\"")
						util.WriteToOut(out, "Updated system prompt\n")
						continue
					}

					if prompt == "/help" {
						util.WriteToOut(out, "Commands:\n")
						util.WriteToOut(out, "  /bye, /exit, /quit - Exit the chat\n")
						util.WriteToOut(out, "  /parameters - Show current model parameters\n")
						util.WriteToOut(out, "  /reset, /clear - Reset chat context\n")
						util.WriteToOut(out, "  /set <name> <value> - Set a model parameter\n")
						util.WriteToOut(out, "  /system-prompt <prompt> - Set the system prompt\n")
						util.WriteToOut(out, "  /help - Show this help message\n")
						continue
					}

					util.WriteToOut(out, "Unknown command '"+prompt+"'. See /help for supported commands.\n")
					continue
				}

				conversation.AddMessage(azure_models.ChatMessageRoleUser, prompt)

				req := azure_models.ChatCompletionOptions{
					Messages: conversation.GetMessages(),
					Model:    modelName,
				}

				mp.UpdateRequest(&req)

				sp := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(errOut))
				sp.Start()
				defer sp.Stop()

				resp, err := client.GetChatCompletionStream(req)
				if err != nil {
					return err
				}

				defer resp.Reader.Close()

				messageBuilder := strings.Builder{}

				for {
					completion, err := resp.Reader.Read()
					if err != nil {
						if errors.Is(err, io.EOF) {
							break
						}
						return err
					}

					sp.Stop()

					for _, choice := range completion.Choices {
						// Streamed responses from the OpenAI API have their data in `.Delta`, while
						// non-streamed responses use `.Message`, so let's support both
						if choice.Delta != nil && choice.Delta.Content != nil {
							content := choice.Delta.Content
							messageBuilder.WriteString(*content)
							util.WriteToOut(out, *content)
						} else if choice.Message != nil && choice.Message.Content != nil {
							content := choice.Message.Content
							messageBuilder.WriteString(*content)
							util.WriteToOut(out, *content)
						}

						// Introduce a small delay in between response tokens to better simulate a conversation
						if terminal.IsTerminalOutput() {
							time.Sleep(10 * time.Millisecond)
						}
					}
				}

				util.WriteToOut(out, "\n")
				messageBuilder.WriteString("\n")

				conversation.AddMessage(azure_models.ChatMessageRoleAssistant, messageBuilder.String())

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
