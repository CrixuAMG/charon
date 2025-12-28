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
	fmt.Println("Usage: charon <command>\n")

	var names []string
	for name := range c.router.All() {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		fmt.Println(c.router.All()[name].Help())
	}
	return nil
}

