// Package cmds provides the implementation backing token-forge's cli.
// This section of the cmds package holds the implementation for the login
// command.
package cmds

import (
	"fmt"
	"sync"

	"github.com/pyqlsa/token-forge/internal/ghtoken"
)

// tokenSource is an interface to a source of tokens.
type tokenSource interface {
	pop() (*ghtoken.GhToken, error)
	done() bool
	remaining() uint64
}

// gTokenSource is a tokenSource that supplies randomly generated tokens.
type gTokenSource struct {
	mu        sync.Mutex
	remain    uint64
	tokenFunc func() *ghtoken.GhToken
}

// done signifies if the token source has been exhausted.
func (g *gTokenSource) done() bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.remain < 1
}

// remaining returns the number of tokens that remain in the source.
func (g *gTokenSource) remaining() uint64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.remain
}

// pop removes and returns a token from the source, or returns an error if the
// source has been exhausted.
func (g *gTokenSource) pop() (*ghtoken.GhToken, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.remain < 1 {
		return nil, fmt.Errorf("token source has been drained")
	}
	g.remain--

	return g.tokenFunc(), nil
}

// generatedTokenSource returns a gTokenSource that supplies tokens with the
// given prefix, up to the given limit.
func generatedTokenSource(prefix string, limit uint64) *gTokenSource {
	//nolint:exhaustruct
	return &gTokenSource{
		tokenFunc: GenGhTokenFunc(prefix),
		remain:    limit,
	}
}

type nTokenSource struct {
	mu     sync.Mutex
	remain uint64
}

// done signifies if the token source has been exhausted.
func (n *nTokenSource) done() bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.remain < 1
}

// remaining returns the number of tokens that remain in the source.
func (n *nTokenSource) remaining() uint64 {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.remain
}

// pop removes and returns a token from the source, or returns an error if the
// source has been exhausted.
func (n *nTokenSource) pop() (*ghtoken.GhToken, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.remain < 1 {
		return nil, fmt.Errorf("token source has been drained")
	}
	n.remain--

	return nil, nil //nolint:nilnil
}

func nilTokenSource(limit uint64) *nTokenSource {
	//nolint:exhaustruct
	return &nTokenSource{
		remain: limit,
	}
}

// sTokenSource is a tokenSource that supplies tokens that were acquired from
// some other static source.
type sTokenSource struct {
	mu     sync.Mutex
	tokens []*ghtoken.GhToken
}

// done signifies if the token source has been exhausted.
func (s *sTokenSource) done() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.tokens) < 1
}

// remaining returns the number of tokens that remain in the source.
func (s *sTokenSource) remaining() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	return uint64(len(s.tokens))
}

// pop removes and returns a token from the source, or returns an error if the
// source has been exhausted.
func (s *sTokenSource) pop() (*ghtoken.GhToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.tokens) < 1 {
		return nil, fmt.Errorf("token source has been drained")
	}
	n := len(s.tokens) - 1
	tok := s.tokens[n]
	s.tokens = s.tokens[:n]

	return tok, nil
}

// fileTokenSource returns an sTokenSource, saturated with tokens from the
// given file, up to the given limit.
func fileTokenSource(file string, limit uint64) (*sTokenSource, error) {
	toks, err := getNumTokensFromFile(file, limit)
	if err != nil {
		return nil, fmt.Errorf("failed getting tokens from file '%s': %w", file, err)
	}

	//nolint:exhaustruct
	return &sTokenSource{
		tokens: toks,
	}, nil
}
