// Package cmds provides the implementation backing token-forge's cli.
// This section of the cmds package holds the implementation for the disect
// command.
package cmds

import (
	"fmt"
	"log"

	"github.com/pyqlsa/token-forge/internal/fileutil"
	"github.com/pyqlsa/token-forge/internal/ghtoken"
)

// DisectCmd represents the disect tokens cli command.
type DisectCmd struct {
	Globals
	TokenSourceArgs
	TokenParams
}

// Run the disect tokens command to inspect GitHub tokens.
func (d *DisectCmd) Run() error {
	switch {
	case len(d.Token) > 0:
		token := ghtoken.ParseGhToken(d.Token)
		token.PrintTokenAttributes()

		return nil
	case len(d.File) > 0:
		if err := disectTokensFromFile(d.File); err != nil {
			return fmt.Errorf("failed inspecting tokens: %w", err)
		}

		return nil
	case d.Generated:
		if err := disectGeneratedTokens(d.Prefix, d.NumTokens); err != nil {
			return fmt.Errorf("failed inspecting generated tokens: %w", err)
		}

		return nil
	case d.NoAuth:
		log.Println("[WARNING] the no-auth flag has no effect in this mode, there's no token to inspect!")

		return nil
	default:
		log.Println("[WARNING] we somehow have an unhandled case, please report this!")

		return nil
	}
}

// Print out attributes of tokens in the given file.
func disectTokensFromFile(file string) error {
	tokens, err := fileutil.ReadLines(file)
	if err != nil {
		return fmt.Errorf("failed reading file '%s': %w", file, err)
	}

	for _, tok := range tokens {
		token := ghtoken.ParseGhToken(tok)
		token.PrintTokenAttributes()
	}

	return nil
}

// Print out attributes of generated tokens.
func disectGeneratedTokens(prefix string, numTokens uint64) error {
	if len(prefix) > 0 && !ghtoken.IsValidPrefix(prefix) {
		return fmt.Errorf("prefix '%s' is not a valid token prefix", prefix)
	}

	genToken := GenGhTokenFunc(prefix)
	for i := uint64(0); i < numTokens; i++ {
		fake := genToken()
		fake.PrintTokenAttributes()
	}

	return nil
}
