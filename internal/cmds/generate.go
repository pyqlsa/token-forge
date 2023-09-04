// Package cmds provides the implementation backing token-forge's cli.
// This section of the cmds package hold the implementation for the generate
// command.
package cmds

import (
	"fmt"

	"github.com/pyqlsa/token-forge/internal/datautil"
	"github.com/pyqlsa/token-forge/internal/ghtoken"
)

// GenCmd represents the generate tokens cli command.
type GenCmd struct {
	Globals
	TokenParams
}

// Run the generate tokens command to generate GitHub-like tokens.
func (d *GenCmd) Run() error {
	if len(d.Prefix) > 0 && !ghtoken.IsValidPrefix(d.Prefix) {
		return fmt.Errorf("prefix '%s' is not a valid token prefix", d.Prefix)
	}

	genToken := GenGhTokenFunc(d.Prefix)
	for i := uint64(0); i < d.NumTokens; i++ {
		fake := genToken()
		fmt.Println(fake.FullToken)
	}

	return nil
}

// GenGhTokenFunc returns a token generation function based on the provided
// token prefix; if an empty string is provided, the returned function will
// generate tokens with a randomly selected prefix; if a non-empty string is
// provided, then the returned function will generate all tokens w/ the given
// prefix. This does not check the validity of the provided token prefix.
func GenGhTokenFunc(prefix string) func() *ghtoken.GhToken {
	if len(prefix) > 0 {
		return func() *ghtoken.GhToken {
			return genGhTokenWithPrefix(prefix)
		}
	}

	return genGhTokenRandPrefix
}

// Generate a GitHub-like token w/ the given prefix.
func genGhTokenWithPrefix(prefix string) *ghtoken.GhToken {
	var input, crc string
	// Generate token components and 0-pad until desired length is achieved.
	// When length is exceeded, try again.
	for len(input) != ghtoken.InputLength || len(crc) != ghtoken.ChecksumLength {
		// 22 bytes often encodes to 30 character strings, but sometimes underflows
		// to 29 character strings; 23 bytes often encodes to 30 character strings,
		// but sometimes overflows to 31 character strings.
		// From a manual analysis compared against a small set of real GitHub tokens,
		// it appears that generating 23 bytes most closely mimics real token data;
		// Larger scale automated analysis would better serve to prove this hypothesis.
		input = datautil.EncodeBase62(datautil.GenerateSecureRandomBytes(datautil.SelectInsecureRandomInt(22, 23)), ghtoken.Base62Alphabet)
		for len(input) < ghtoken.InputLength {
			input = fmt.Sprintf("0%s", input) // unsure if we should 0-pad token, but doing it anyways
		}
		crc = datautil.EncodeBase62(datautil.Crc32ChecksumBytes(input), ghtoken.Base62Alphabet)
		for len(crc) < ghtoken.ChecksumLength {
			crc = fmt.Sprintf("0%s", crc)
		}
	}

	payload := fmt.Sprintf("%s%s", input, crc)
	fullToken := fmt.Sprintf("%s%s%s", prefix, ghtoken.Sep, payload)

	return ghtoken.ParseGhToken(fullToken)
}

// Generate a GitHub-like token w/ a randomly selected prefix.
func genGhTokenRandPrefix() *ghtoken.GhToken {
	return genGhTokenWithPrefix(datautil.SelectInsecureRandomStr(ghtoken.GetValidPrefixes()...))
}
