package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/crixuamg/charon/internal/app"
)

type ManCommand struct{}

func NewMan() *ManCommand {
	return &ManCommand{}
}

func (c *ManCommand) Name() string { return "man" }

func (c *ManCommand) Help() string {
	return "man             Print a man page for charon to stdout"
}

// Run writes a roff-formatted man page to stdout.
// Redirect to a file or pipe through man -l - to read it.
func (c *ManCommand) Run(_ *app.Context, _ []string) error {
	date := time.Now().Format("January 2006")
	page := buildManPage(date)
	_, err := fmt.Fprint(os.Stdout, page)
	return err
}

func buildManPage(date string) string {
	var b strings.Builder

	header(&b, date)

	section(&b, "NAME")
	b.WriteString(".PP\ncharon \\- project navigator and launcher\n")

	section(&b, "SYNOPSIS")
	b.WriteString(".PP\n.B charon\n[\\-\\-config\\ \\fIFILE\\fP]\\ [\\fICOMMAND\\fP]\\ [\\fIARGS\\fP]\n")

	section(&b, "DESCRIPTION")
	b.WriteString(".PP\nCharon is an interactive terminal UI for managing and opening development\n")
	b.WriteString("projects. It integrates with the Kitty terminal emulator to open project\n")
	b.WriteString("workspaces in new tabs.\n")

	section(&b, "OPTIONS")
	flag(&b, "--config FILE", "Path to configuration file. Overrides the CHARON_CONFIG environment variable and the default location (~/.config/charon/.charon.yaml).")

	section(&b, "COMMANDS")
	cmd(&b, "tui", "Start the interactive project list (default when no command is given).")
	cmd(&b, "open <project>", "Open a project directly by name or alias without entering the TUI.")
	cmd(&b, "add <path>", "Add the project at PATH to the configuration.")
	cmd(&b, "edit", "Open the configuration file in $EDITOR.")
	cmd(&b, "export-aliases", "Print shell functions for all projects to stdout. Redirect to .zshrc or .bashrc and source the file to activate short aliases.")
	cmd(&b, "man", "Print this man page to stdout.")
	cmd(&b, "help", "List available commands.")

	section(&b, "CONFIGURATION")
	b.WriteString(".PP\nThe configuration file is a YAML document located at\n")
	b.WriteString(".B ~/.config/charon/.charon.yaml\n")
	b.WriteString("by default. The path can be overridden with the\n")
	b.WriteString(".B CHARON_CONFIG\nenvironment variable or the\n")
	b.WriteString(".B --config\nflag.\n")
	b.WriteString(".PP\nExample:\n")
	b.WriteString(".nf\n")
	b.WriteString("projects:\n")
	b.WriteString("  - name: my-api\n")
	b.WriteString("    alias: api\n")
	b.WriteString("    path: ~/code/my-api\n")
	b.WriteString("    tags: [work, backend]\n")
	b.WriteString("    env:\n")
	b.WriteString("      PORT: \"3001\"\n")
	b.WriteString("    hooks:\n")
	b.WriteString("      before: docker compose up -d db\n")
	b.WriteString("    tasks:\n")
	b.WriteString("      - nvim\n")
	b.WriteString("      - yarn dev\n")
	b.WriteString(".fi\n")

	section(&b, "ENVIRONMENT")
	envVar(&b, "CHARON_CONFIG", "Override the default config file path.")
	envVar(&b, "EDITOR", "Editor used by the 'edit' command and the note editor.")
	envVar(&b, "KITTY_LISTEN_ON", "Kitty remote control socket. Detected automatically if not set.")

	section(&b, "KEY BINDINGS")
	b.WriteString(".PP\nThe following keys are available in the project list:\n")
	b.WriteString(".TS\nallbox;\nl l.\n")
	keyRow(&b, "↑ / k", "Move cursor up")
	keyRow(&b, "↓ / j", "Move cursor down")
	keyRow(&b, "enter", "Open selected project")
	keyRow(&b, "space", "Toggle multi-select")
	keyRow(&b, "/", "Enter search mode")
	keyRow(&b, "s", "Cycle sort mode")
	keyRow(&b, "f", "Cycle filter mode")
	keyRow(&b, "T", "Cycle tag filter")
	keyRow(&b, "l", "Cycle layout")
	keyRow(&b, "a", "Add project")
	keyRow(&b, "e", "Edit project")
	keyRow(&b, "d / x", "Delete project")
	keyRow(&b, "A", "Archive / unarchive project")
	keyRow(&b, "ctrl+a", "Toggle show archived")
	keyRow(&b, "n", "View note")
	keyRow(&b, "ctrl+n", "Edit note in $EDITOR")
	keyRow(&b, "ctrl+o", "Reveal path in file manager")
	keyRow(&b, "ctrl+d", "Bulk delete selected")
	keyRow(&b, "ctrl+r", "Bulk archive selected")
	keyRow(&b, "t", "Manage task sets")
	keyRow(&b, "g / G", "Jump to first / last")
	keyRow(&b, "X", "Cycle theme")
	keyRow(&b, "q / ctrl+c", "Quit")
	b.WriteString(".TE\n")

	section(&b, "FILES")
	b.WriteString(".TP\n.B ~/.config/charon/.charon.yaml\nDefault configuration file.\n")
	b.WriteString(".TP\n.B ~/.local/share/charon/charon.db\nSQLite database tracking project access times and counts.\n")

	section(&b, "SEE ALSO")
	b.WriteString(".PP\nkitty(1)\n")

	return b.String()
}

func header(b *strings.Builder, date string) {
	b.WriteString(fmt.Sprintf(".TH CHARON 1 \"%s\" \"Charon\" \"User Commands\"\n", date))
}

func section(b *strings.Builder, name string) {
	b.WriteString(fmt.Sprintf(".SH %s\n", name))
}

func flag(b *strings.Builder, name, desc string) {
	b.WriteString(fmt.Sprintf(".TP\n.B %s\n%s\n", name, desc))
}

func cmd(b *strings.Builder, name, desc string) {
	b.WriteString(fmt.Sprintf(".TP\n.B %s\n%s\n", name, desc))
}

func envVar(b *strings.Builder, name, desc string) {
	b.WriteString(fmt.Sprintf(".TP\n.B %s\n%s\n", name, desc))
}

func keyRow(b *strings.Builder, key, desc string) {
	b.WriteString(fmt.Sprintf("%s\t%s\n", key, desc))
}
