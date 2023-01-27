package ui

import (
	"testing"

	"github.com/mitchellh/cli"
	"github.com/shoenig/test/must"
)

func TestColoredUI(t *testing.T) {
	mUI := cli.NewMockUi()
	ui := ColoredUi{
		OutputColor: UiColorBlue,
		InfoColor:   UiColorCyan,
		WarnColor:   UiColorYellow,
		ErrorColor:  UiColorRed,
		ColorWhen:   UiColorAlways,
		Ui:          mUI,
	}
	ui.Output("test")
	must.Eq(t, mUI.OutputWriter.String(), "\x1b[94mtest\x1b[0m\n")
	mUI.OutputWriter.Reset()

	ui.ColorWhen = UiColorNever
	ui.Output("test")
	must.Eq(t, mUI.OutputWriter.String(), "test\n")
	mUI.OutputWriter.Reset()

	// Auto will test in the negative because it doesn't use a tty
	ui.ColorWhen = UiColorAuto
	ui.Output("test")
	must.Eq(t, mUI.OutputWriter.String(), "test\n")
}
