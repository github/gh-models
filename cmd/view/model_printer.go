package view

import (
	"strings"

	"github.com/cli/go-gh/v2/pkg/tableprinter"
	"github.com/cli/go-gh/v2/pkg/term"
	"github.com/github/gh-models/internal/azure_models"
	"github.com/mgutz/ansi"
)

var (
	lightGrayUnderline = ansi.ColorFunc("white+du")
)

type modelPrinter struct {
	modelSummary *azure_models.ModelSummary
	modelDetails *azure_models.ModelDetails
	printer      tableprinter.TablePrinter
}

func newModelPrinter(summary *azure_models.ModelSummary, details *azure_models.ModelDetails, terminal term.Term) modelPrinter {
	width, _, _ := terminal.Size()
	printer := tableprinter.New(terminal.Out(), terminal.IsTerminalOutput(), width)
	return modelPrinter{modelSummary: summary, modelDetails: details, printer: printer}
}

func (p *modelPrinter) render() error {
	modelSummary := p.modelSummary
	if modelSummary != nil {
		p.printLabelledLine("Display name:", modelSummary.FriendlyName)
		p.printLabelledLine("Summary name:", modelSummary.Name)
		p.printLabelledLine("Publisher:", modelSummary.Publisher)
		p.printLabelledLine("Summary:", modelSummary.Summary)
	}

	modelDetails := p.modelDetails
	if modelDetails != nil {
		p.printLabelledLine("Tags:", strings.Join(modelDetails.Tags, ", "))
		p.printLabelledLine("Supported input types:", strings.Join(modelDetails.SupportedInputModalities, ", "))
		p.printLabelledLine("Supported output types:", strings.Join(modelDetails.SupportedOutputModalities, ", "))
		p.printLabelledLine("License:", modelDetails.License)
		p.printMultipleLinesWithLabel("License description:", modelDetails.LicenseDescription)
		p.printMultipleLinesWithLabel("Description:", modelDetails.Description)
		p.printMultipleLinesWithLabel("Notes:", modelDetails.Notes)
	}

	err := p.printer.Render()
	if err != nil {
		return err
	}

	return nil
}

func (p *modelPrinter) printLabelledLine(label string, value string) {
	if value == "" {
		return
	}
	p.addLabel(label)
	p.printer.AddField(value)
	p.printer.EndRow()
}

func (p *modelPrinter) printMultipleLinesWithLabel(label string, value string) {
	if value == "" {
		return
	}
	p.addLabel(label)
	p.printer.AddField(value, tableprinter.WithTruncate(nil))
	p.printer.EndRow()
	p.printBlankLine()
}

func (p *modelPrinter) printBlankLine() {
	p.printer.AddField("")
	p.printer.EndRow()
}

func (p *modelPrinter) addLabel(label string) {
	p.printer.AddField(label, tableprinter.WithTruncate(nil), tableprinter.WithColor(lightGrayUnderline))
}
