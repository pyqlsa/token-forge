// Package main provides entrance for cli.
package main

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/pyqlsa/token-forge/internal/cmds"
)

var (
	version string // populated by build script.
	commit  string // populated by build script.
)

// VersionCmd represent version cli command.
type VersionCmd struct{}

// Run version command.
func (v VersionCmd) Run() error {
	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Commit: %s\n", commit)

	return nil
}

var cli struct {
	Version  VersionCmd     `cmd:""        help:"Print version and exit."`
	Generate cmds.GenCmd    `aliases:"gen" cmd:""                                     help:"Generate GitHub-like tokens."`
	Disect   cmds.DisectCmd `aliases:"dis" cmd:""                                     help:"Disect GitHub-like tokens."`
	Login    cmds.LoginCmd  `cmd:""        help:"Test login with one or more tokens."`
	Local    cmds.LocalCmd  `cmd:""        help:"Perform a local collision test."`
	IPCheck  cmds.IPCmd     `aliases:"ip"  cmd:""                                     help:"Check resolved public ip address."`
}

// Main.
func main() {
	ctx := kong.Parse(&cli,
		kong.Name("token-forge"),
		kong.Description("A tool to 'work' with GitHub tokens."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			NoAppSummary:        false,
			Indenter:            kong.SpaceIndenter,
			Compact:             false,
			Summary:             true,
			NoExpandSubcommands: false,
			FlagsLast:           true,
			Tree:                true,
			WrapUpperBound:      120,
		}))
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
