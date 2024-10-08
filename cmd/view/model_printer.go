package view

import (
	"github.com/cli/go-gh/v2/pkg/tableprinter"
	"github.com/cli/go-gh/v2/pkg/term"
	"github.com/github/gh-models/internal/azure_models"
	"github.com/github/gh-models/pkg/util"
)

type modelPrinter struct {
	model   *azure_models.ModelSummary
	printer tableprinter.TablePrinter
}

func newModelPrinter(model *azure_models.ModelSummary, terminal term.Term) modelPrinter {
	width, _, _ := terminal.Size()
	printer := tableprinter.New(terminal.Out(), terminal.IsTerminalOutput(), width)
	return modelPrinter{model: model, printer: printer}
}

func (p *modelPrinter) render() error {
	p.addLabeledValue("Display Name", p.model.FriendlyName)
	p.addLabeledValue("Model Name", p.model.Name)

	err := p.printer.Render()
	if err != nil {
		return err
	}

	return nil
}

func (p *modelPrinter) addLabeledValue(label string, value string) {
	p.printer.AddField(label, tableprinter.WithColor(util.LightGrayUnderline))
	p.printer.AddField(value)
	p.printer.EndRow()
}
