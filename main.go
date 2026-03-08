package main

import (
	"flag"
	"log"

	"github.com/crixuamg/charon/internal/app"
	"github.com/crixuamg/charon/internal/commands"
	"github.com/crixuamg/charon/internal/config"
)

func main() {
	log.SetPrefix("charon: ")
	log.SetFlags(0)

	configFlag := flag.String("config", "", "path to config file (overrides CHARON_CONFIG and default location)")
	flag.Parse()

	cfg, err := config.LoadFrom(*configFlag)
	if err != nil {
		log.Fatal(err)
	}

	if err := app.AutoDiscover(cfg); err != nil {
		log.Printf("warning: autodiscover: %v", err)
	}
	app.UpdateProjectStatus(cfg)

	for _, w := range cfg.Validate() {
		log.Printf("warning: %s", w)
	}

	ctx := &app.Context{
		Config: cfg,
	}

	router := commands.NewRouter("tui")

	router.Add(commands.NewOpen())
	router.Add(commands.NewTUI())
	router.Add(commands.NewHelp(router))
	router.Add(commands.NewEdit())
	router.Add(commands.NewAdd())
	router.Add(commands.NewExportAliases())
	router.Add(commands.NewMan())

	if err := router.Run(ctx, flag.Args()); err != nil {
		log.Fatal(err)
	}
}
