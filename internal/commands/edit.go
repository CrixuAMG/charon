package commands

import (
	"os"
	"os/exec"

	"github.com/crixuamg/charon/internal/app"
	"github.com/crixuamg/charon/internal/config"
)

type EditCommand struct{}

func NewEdit() *EditCommand {
	return &EditCommand{}
}

func (c *EditCommand) Name() string { return "edit" }

func (c *EditCommand) Help() string {
	return "edit            Edit Charon's config file"
}

func (c *EditCommand) Run(_ *app.Context, _ []string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	configPath, err := config.ConfigPath()
	if err != nil {
		return err
	}

	cmd := exec.Command(editor, configPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
