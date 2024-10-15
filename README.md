## GitHub Models extension

Use the GitHub Models service from the CLI!

## Using

### Prerequisites

The extension requires the [`gh` CLI](https://cli.github.com/) to be installed and in the `PATH`. The extension also requires the user have authenticated via `gh auth`.

### Installing

After installing the `gh` CLI, from a command-line run:
```shell
gh extension install https://github.com/github/gh-models
```

### Examples

#### Listing models

```shell
gh models list
```

Example output:
```shell
Name                          Friendly Name                 Publisher
AI21-Jamba-Instruct           AI21-Jamba-Instruct           AI21 Labs
gpt-4o                        OpenAI GPT-4o                 Azure OpenAI Service
gpt-4o-mini                   OpenAI GPT-4o mini            Azure OpenAI Service
Cohere-command-r              Cohere Command R              cohere
Cohere-command-r-plus         Cohere Command R+             cohere
```

Use the value in the "Name" column when specifying the model on the command-line.

#### Running inference

##### REPL mode

Run the extension in REPL mode. This will prompt you for which model to use.
```shell
gh models run
```

In REPL mode, use `/help` to list available commands. Otherwise just type your prompt and hit ENTER to send to the model.

##### Single-shot mode

Run the extension in single-shot mode. This will print the model output and exit.
```shell
gh models run gpt-4o-mini "why is the sky blue?"
```

Run the extension with output from a command. This uses single-shot mode.
```shell
cat README.md | gh models run gpt-4o-mini "summarize this text"
```
