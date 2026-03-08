package commands

import (
	"fmt"
	"strings"

	"github.com/crixuamg/charon/internal/app"
)

type ExportAliasesCommand struct{}

func NewExportAliases() *ExportAliasesCommand {
	return &ExportAliasesCommand{}
}

func (c *ExportAliasesCommand) Name() string { return "export-aliases" }

func (c *ExportAliasesCommand) Help() string {
	return "export-aliases  Print shell alias functions for all projects"
}

// Run writes one shell function per project (and one for each alias) to stdout.
// Redirect output to ~/.zshrc or ~/.bashrc and source it to activate.
func (c *ExportAliasesCommand) Run(ctx *app.Context, _ []string) error {
	var b strings.Builder

	for _, p := range ctx.Config.Projects {
		name := p.Name
		// Generate a shell-safe function name: replace non-alphanumeric chars with _.
		funcName := shellSafeName(name)
		b.WriteString(fmt.Sprintf("function %s() { charon open %q; }\n", funcName, name))

		// If the project has a short alias, also emit a function for that.
		if p.Alias != "" {
			aliasFuncName := shellSafeName(p.Alias)
			if aliasFuncName != funcName {
				b.WriteString(fmt.Sprintf("function %s() { charon open %q; }\n", aliasFuncName, name))
			}
		}
	}

	fmt.Print(b.String())
	return nil
}

// shellSafeName converts a project name to a valid shell identifier.
func shellSafeName(name string) string {
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}
