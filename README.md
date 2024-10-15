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

## Notice

Remember when interacting with a model you are experimenting with AI, so content mistakes are possible. The feature is
subject to various limits (including requests per minute, requests per day, tokens per request, and concurrent requests)
and is not designed for production use cases. GitHub Models uses
[Azure AI Content Safety](https://azure.microsoft.com/en-us/products/ai-services/ai-content-safety). These filters
cannot be turned off as part of the GitHub Models experience. If you decide to employ models through a paid service,
please configure your content filters to meet your requirements. This service is under
[GitHub's Pre-release Terms](https://docs.github.com/en/site-policy/github-terms/github-pre-release-license-terms). Your
use of the GitHub Models is subject to the following
[Product Terms](https://www.microsoft.com/licensing/terms/productoffering/MicrosoftAzure/allprograms) and
[Privacy Statement](https://www.microsoft.com/licensing/terms/product/PrivacyandSecurityTerms/MCA). Content within this
Repository may be subject to additional license terms.
