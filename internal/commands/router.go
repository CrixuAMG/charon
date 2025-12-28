package commands

import (
	"fmt"

	"github.com/crixuamg/charon/internal/app"
)

type Router struct {
	commands   map[string]Command
	defaultCmd string
}

func NewRouter(defaultCmd string) *Router {
	return &Router{
		commands:   make(map[string]Command),
		defaultCmd: defaultCmd,
	}
}

func (r *Router) Add(cmd Command) {
	r.commands[cmd.Name()] = cmd
}

func (r *Router) Run(ctx *app.Context, args []string) error {
	if len(args) == 0 {
		if cmd, ok := r.commands[r.defaultCmd]; ok {
			return cmd.Run(ctx, nil)
		}
		return fmt.Errorf("no default command configured")
	}

	cmd, ok := r.commands[args[0]]
	if !ok {
		return fmt.Errorf("unknown command: %s", args[0])
	}

	return cmd.Run(ctx, args[1:])
}

func (r *Router) All() map[string]Command {
	return r.commands
}
