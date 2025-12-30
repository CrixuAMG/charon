package commands

import (
	"fmt"

	"github.com/crixuamg/charon/internal/app"
	"github.com/crixuamg/charon/internal/config"
	"github.com/crixuamg/charon/internal/kitty"
)

type OpenCommand struct {
	cfg config.Config
}

func NewOpen() *OpenCommand {
	return &OpenCommand{}
}

func (c *OpenCommand) Name() string { return "open" }

func (c *OpenCommand) Help() string {
	return "open <project>  Open a project and exit"
}

func (c *OpenCommand) Run(ctx *app.Context, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: %s", c.Help())
	}

	project, err := ctx.Config.FindProject(args[0])
	if err != nil {
		return err
	}

	return kitty.OpenProject(*project, ctx.Config)
}
