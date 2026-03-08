package commands

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/crixuamg/charon/internal/app"
	"github.com/crixuamg/charon/internal/tui"
)

type TUICommand struct{}

func NewTUI() *TUICommand {
	return &TUICommand{}
}

func (c *TUICommand) Name() string { return "tui" }

func (c *TUICommand) Help() string {
	return "tui             Start interactive UI (default)"
}

func (c *TUICommand) Run(ctx *app.Context, _ []string) error {
	model := tui.NewModel(ctx.Config)
	defer model.Cleanup()

	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
