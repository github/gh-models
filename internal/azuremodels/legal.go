package azuremodels

import (
	"github.com/cli/go-gh/v2/pkg/term"
	"github.com/github/gh-models/pkg/util"
	"github.com/mgutz/ansi"
)

const notice = "ℹ︎ Azure hosted. AI powered, can make mistakes. Not intended for production/sensitive data.\nFor more information, see https://ai.azure.com/github/model/docs"

func LegalNotice() {
	msg := notice
	if !term.IsColorDisabled() || term.IsColorForced() {
		msg = ansi.Color(msg, "yellow")
	}
	util.WriteToOut(term.FromEnv().ErrOut(), msg+"\n")
}
