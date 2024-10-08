package view

import (
	"github.com/cli/go-gh/v2/pkg/tableprinter"
	"github.com/cli/go-gh/v2/pkg/term"
	"github.com/github/gh-models/internal/azure_models"
	"github.com/github/gh-models/pkg/util"
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
		p.addLabeledValue("Display name:", modelSummary.FriendlyName)
		p.addLabeledValue("Summary name:", modelSummary.Name)
		p.addLabeledValue("Publisher:", modelSummary.Publisher)
		p.addLabeledValue("Summary:", modelSummary.Summary)
	}

	modelDetails := p.modelDetails
	if modelDetails != nil {
		p.addLabel("Description:")
		p.printer.AddField(modelDetails.Description, tableprinter.WithTruncate(nil))
		p.printer.EndRow()
	}

	err := p.printer.Render()
	if err != nil {
		return err
	}

	return nil
}

func (p *modelPrinter) addLabel(label string) {
	p.printer.AddField(label, tableprinter.WithTruncate(nil), tableprinter.WithColor(util.LightGrayUnderline))
}

func (p *modelPrinter) addLabeledValue(label string, value string) {
	p.addLabel(label)
	p.printer.AddField(value)
	p.printer.EndRow()
}
