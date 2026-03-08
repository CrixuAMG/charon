package commands

import (
	"fmt"
	"sort"

	"github.com/crixuamg/charon/internal/app"
)

type HelpCommand struct {
	router *Router
}

func NewHelp(router *Router) *HelpCommand {
	return &HelpCommand{router: router}
}

func (c *HelpCommand) Name() string { return "help" }

func (c *HelpCommand) Help() string {
	return "help            Show this help"
}

func (c *HelpCommand) Run(_ *app.Context, _ []string) error {
	fmt.Println("Usage: charon <command>")

	cmds := c.router.All()
	names := make([]string, 0, len(cmds))
	for name := range cmds {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		fmt.Println(cmds[name].Help())
	}
	return nil
}
