package main

import (
	"log"
	"os"

	"github.com/crixuamg/charon/internal/app"
	"github.com/crixuamg/charon/internal/commands"
	"github.com/crixuamg/charon/internal/config"
)

func main() {
	log.SetPrefix("charon: ")
	log.SetFlags(0)

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := &app.Context{
		Config: cfg,
	}

	router := commands.NewRouter("tui")

	router.Add(commands.NewOpen())
	router.Add(commands.NewTUI())
	router.Add(commands.NewHelp(router))

	if err := router.Run(ctx, os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
