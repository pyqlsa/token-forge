// Package cmds provides the implementation backing token-forge's cli.
// This section of the cmds package holds implementation agnostic things used
// in other sections of the cmds package.
package cmds

import (
	"fmt"
	"log"
	"math"
	"net/url"
	"os"

	"github.com/pyqlsa/token-forge/internal/fileutil"
	"github.com/pyqlsa/token-forge/internal/ghtoken"
)

// Globals represents globally shared flags for cli commands.
type Globals struct {
	Debug bool `help:"Enable debug mode"`
}

// TokenSourceArgs represents a source of tokens.
type TokenSourceArgs struct {
	Token     string `group:"Source" help:"Token to use."                                                         required:"" short:"t" xor:"source"`
	File      string `group:"Source" help:"Path to file with tokens."                                             required:"" short:"f" type:"existingfile" xor:"source"`
	Generated bool   `group:"Source" help:"Use one or more generated tokens."                                     required:"" short:"g" xor:"source"`
	NoAuth    bool   `group:"Source" help:"Simply interact w/ the rate limit api with an unauthenticated client." required:"" short:"x" xor:"source"`
}

// TokenParams represents parameters for token generation.
type TokenParams struct {
	BatchSize int    `default:"1000"       group:"Token Params"                                                                                                                                help:"When testing for collisions, the number of tokens to test concurrently." short:"b"`
	NumTokens uint64 `default:"1"          group:"Token Params"                                                                                                                                help:"Number of tokens to test."                                               short:"n"` // max = 18446744073709551615`
	Prefix    string `group:"Token Params" help:"Token prefix to use; if not specified, each generated token will have a randomly selected prefix; only has an effect when generating tokens." short:"p"`
}

// ProxyConfig represents parameters for setting a proxy.
type ProxyConfig struct {
	Proxy string `group:"Proxy Config" help:"Proxy to use for outbound connections."`
}

// setProxy validates a url string and sets it as a proxy via environment
// variables; if the string is a valid url, it unsets HTTP_PROXY, HTTPS_PROXY,
// and NO_PROXY, then sets HTTP_PROXY and HTTPS_PROXY. These variables are
// typically picked up by default http client implementations in underlying
// libraries.
// do this instead? https://medium.com/@prasincs/proxy-aware-http-client-in-go-6a3a487cbe5b
func setProxy(u string) error {
	if u == "" {
		return nil
	}

	proxy, err := url.Parse(u)
	if err != nil {
		return fmt.Errorf("failed parsing proxy url: %w", err)
	}

	uvars := []string{"HTTP_PROXY", "http_proxy", "HTTPS_PROXY", "https_proxy", "NO_PROXY", "no_proxy"}
	for _, u := range uvars {
		if err := os.Unsetenv(u); err != nil {
			return fmt.Errorf("failed unsetting %s in prep for configuring proxy: %w", u, err)
		}
	}

	svars := []string{"HTTP_PROXY", "HTTPS_PROXY"}
	for _, s := range svars {
		if err := os.Setenv(s, proxy.String()); err != nil {
			return fmt.Errorf("failed setting %s while configuring proxy: %w", s, err)
		}
	}

	return nil
}

// Get the first n tokens from a given file; if n < 1, max uint64 is used;
// malformed tokens are discarded and do not count against the limit.
func getNumTokensFromFile(file string, num uint64) ([]*ghtoken.GhToken, error) {
	if num < 1 {
		num = math.MaxUint64
	}

	tokens := make([]*ghtoken.GhToken, 0)
	toks, err := fileutil.ReadLines(file)
	if err != nil {
		return tokens, fmt.Errorf("failed to get tokens from file '%s': %w", file, err)
	}

	for _, tok := range toks {
		token := ghtoken.ParseGhToken(tok)
		if len(token.FullToken) < 1 {
			log.Printf("error: token '%s' is malformed; skipping...", tok)
		} else {
			tokens = append(tokens, token)
		}
		if uint64(len(tokens)) >= num {
			break
		}
	}
	log.Printf("found %d tokens...", len(tokens))

	return tokens, nil
}
