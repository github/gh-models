package azuremodels

import (
	"github.com/mgutz/ansi"
	"github.com/spf13/cobra"
)

const notice = "ℹ︎ Azure hosted. AI powered, can make mistakes. Not intended for production/sensitive data.\nFor more information, see https://ai.azure.com/github/model/docs"

func LegalNotice(cmd *cobra.Command) {
	cmd.PrintErrln(ansi.Color(notice, "yellow"))
}
