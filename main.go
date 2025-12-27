package main

import (
	"fmt"
	"log"
	"os"

	"github.com/crixuamg/charon/internal/config"
	"github.com/crixuamg/charon/internal/kitty"
	"github.com/crixuamg/charon/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	model := tui.NewModel(cfg)
	defer model.Cleanup()

	// If the user called charon with an argument, like `charon project` we will attempt to open that project or exit with an error
	if len(os.Args) == 2 {
		project := os.Args[1]

		projectCfg, err := cfg.FindProject(project)
		if err != nil {
			log.Fatal(err)
		}

		if err := kitty.OpenProject(*projectCfg, cfg); err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
