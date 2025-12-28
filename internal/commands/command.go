package commands

import "github.com/crixuamg/charon/internal/app"

type Command interface {
	Name() string
	Help() string
	Run(ctx *app.Context, args []string) error
}
