package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/crixuamg/charon/internal/app"
	"github.com/crixuamg/charon/internal/config"
)

type AddCommand struct{}

func NewAdd() *AddCommand {
	return &AddCommand{}
}

func (c *AddCommand) Name() string { return "add" }

func (c *AddCommand) Help() string {
	return "add <path>      Add a project from a directory path"
}

func (c *AddCommand) Run(ctx *app.Context, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: %s", c.Help())
	}

	path, err := filepath.Abs(args[0])
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("path does not exist: %s", path)
	}

	if ctx.Config.HasProjectPath(path) {
		return fmt.Errorf("project at %s is already in the config", path)
	}

	p := config.ProjectFromPath(path)
	ctx.Config.AddProject(p)

	if err := config.Save(ctx.Config); err != nil {
		return err
	}

	fmt.Printf("Added project %q (%s)\n", p.Name, p.Path)
	return nil
}
