// Package ghtoken provides features for working with GitHub tokens.
package ghtoken

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/pyqlsa/token-forge/internal/datautil"
	"github.com/pyqlsa/token-forge/internal/fileutil"
)

// Slice and map const not supported, but treat these like const!
var (
	// Valid PATs (ghp) can be tested by querying user info.
	//
	// Valid oauth tokens (gho) can be tested by querying user info
	// (https://docs.github.com/en/developers/apps/building-oauth-apps/authorizing-oauth-apps#3-use-the-access-token-to-access-the-api)
	// or by searching available installations and repos
	// (https://docs.github.com/en/developers/apps/building-github-apps/identifying-and-authorizing-users-for-github-apps#check-which-installations-resources-a-user-can-access).
	//
	// Valid user-to-server tokens (ghu) can be tested by querying user info or
	// any of the following endpoints (https://docs.github.com/en/developers/apps/building-github-apps/identifying-and-authorizing-users-for-github-apps#user-to-server-requests);
	// apps can be configured to generate expiring user-to-server tokens, and these tokens expire
	// after 8 hours when configured as such (https://docs.github.com/en/developers/apps/building-github-apps/refreshing-user-to-server-access-tokens).
	//
	// Server-to-server / GitHub App tokens (ghs) are nominally generated in the following way
	// (https://docs.github.com/en/rest/reference/apps#create-an-installation-access-token-for-an-app);
	// due to their short validity period (https://docs.github.com/en/developers/apps/building-github-apps/authenticating-with-github-apps#authenticating-as-an-installation),
	// these are probably not that interesting; regardless, the following API
	// endpoints can nominally querying w/ a server-to-server token (https://docs.github.com/en/rest/reference/apps).
	//
	// Refresh tokens (ghr) require an app's client id and client secret to be submitted w/ the
	// refresh token in order to obtain a new user-to-server token (ghu), and they also appear to
	// not quite follow the same format as the other tokens (i.e. longer); thus, refresh tokens
	// don't appear to be that interesting (https://docs.github.com/en/developers/apps/building-github-apps/refreshing-user-to-server-access-tokens#renewing-a-user-token-with-a-refresh-token).

	// Summary per github documentation:
	// ghp meets 36 character format; user-configurable expiration;
	// gho meets 36 character format; no apparent expiration (unless not used within the last year);
	// ghu meets 36 character format; 8 hour expiration (or indefinite, depending on configuration);
	// ghs meets 36 character format; 1 hour expiration;
	// ghr does not meet 36 character format (76 characters); intentionally not included (at the moment).
	prefixes  = []string{"ghp", "gho", "ghu", "ghs"}
	prefixMap = populatePrefixMap()
)

const (
	// Base62Alphabet "enum" to choose a specific base62 alphabet while
	// encoding/decoding; more often than not, due to the alphabet that
	// Golang's big.Int package uses, an inverted alphabet should be chosen
	// when working with tokens produced by or intended to be consumed by
	// GitHub.
	Base62Alphabet = true
	// _32Bit "enum" so we're not using magic numbers; 32 bits = 4 * 1 byte.
	_32Bit = 4
	// PrefixLength "enum" so we're not using magic numbers; prefixes of
	// GitHub tokens are 3 characters long.
	PrefixLength = 3
	// ChecksumLength "enum" so we're not using magic numbers; after the
	// checksum is calculated, it occupies the last 6 characters in the final
	// GitHub token.
	ChecksumLength = 6
	// PayloadLength "enum" so we're not using magic numbers; once the
	// prefix and '_' are stripped from a GitHub token, the resulting payload
	// is 36 characters long.
	PayloadLength = 36
	// InputLength "enum" so we're not using magic numbers; token input is
	// the token payload without the appended checksum.
	InputLength = PayloadLength - ChecksumLength
	// Sep is the character that separates a GitHub token's prefix from
	// it's payload.
	Sep = "_"
)

// Populate the prefix map based on the prefixes slice; this should only need to
// be called once per program execution, and only exists so as to not duplicate
// hard-coded variables.
func populatePrefixMap() map[string]bool {
	pmap := make(map[string]bool)
	for _, p := range prefixes {
		pmap[p] = true
	}

	return pmap
}

// GetValidPrefixes returns a string slice of the valid GitHub token prefixes.
func GetValidPrefixes() []string {
	return prefixes
}

// GhToken holds full token and pre-carved components of a GitHub token.
type GhToken struct {
	FullToken      string `json:"fullToken"`
	Prefix         string `json:"prefix"`
	EncodedPayload string `json:"encodedPayload"`
	EncodedInput   string `json:"encodedInput"`
	EncodedCrc     string `json:"encodedCrc"`
	SchemaChecked  bool   `json:"schemaChecked"`
	SchemaValid    bool   `json:"schemaValid"`
}

// ParseGhToken builds a new ghToken struct from a GitHub token string; if
// token is malformed, a ghToken struct is returned that only has FullToken
// populated (based on the original input).
func ParseGhToken(tok string) *GhToken {
	//nolint:exhaustruct
	token := &GhToken{
		FullToken: strings.TrimSpace(tok),
	}

	token.fillToken()
	token.ValidateSchema()

	return token
}

// fillToken populates the subcomponent struct members based on the current
// contents of FullToken; if the current FullToken member doesn't contain a
// valid separator ('_'), then this function early returns without attempting
// to populate the rest of the struct members; this function does not validate
// correctness of the token, but instead resets the flags denoting any
// attempted validation.
func (token *GhToken) fillToken() {
	token.SchemaValid = false
	token.SchemaChecked = false
	token.FullToken = strings.TrimSpace(token.FullToken)

	prefix, pl, found := strings.Cut(token.FullToken, Sep)
	if !found {
		return
	}

	token.Prefix = prefix
	token.EncodedPayload = pl
	token.EncodedInput = token.EncodedPayload[:len(token.EncodedPayload)-ChecksumLength]
	token.EncodedCrc = token.EncodedPayload[len(token.EncodedPayload)-ChecksumLength:]
}

// IsValidPrefix returns if the given string is a valid ghToken prefix.
func IsValidPrefix(prefix string) bool {
	_, valid := prefixMap[prefix]

	return valid
}

// HasValidPrefix returns whether or not the token's prefix is valid.
func (token GhToken) HasValidPrefix() bool {
	return IsValidPrefix(token.Prefix)
}

// HasValidChecksum a GitHub token's checksum.
func (token GhToken) HasValidChecksum() bool {
	newCrcBytes := datautil.Crc32ChecksumBytes(token.EncodedInput)
	// notes on decoding:
	// https://stackoverflow.com/questions/15848830/decoding-data-from-a-byte-slice-to-uint32
	newCrc := binary.BigEndian.Uint32(newCrcBytes)
	origCrcBytes, ok := datautil.DecodeBase62(token.EncodedCrc, _32Bit, Base62Alphabet)
	if !ok {
		return false
	}

	origCrc := binary.BigEndian.Uint32(origCrcBytes)

	return newCrc == origCrc
}

// ValidateSchema checks the token's prefix and checksum; once a token's schema
// is validated, flags are set in the GhToken to signify that it has been
// checked and whether or not it appears to have a valid schema.
func (token *GhToken) ValidateSchema() {
	token.SchemaChecked = true
	token.SchemaValid = (token.HasValidPrefix() && token.HasValidChecksum())
}

// PrintTokenAttributes prints the ghToken struct, and simply logs an error if
// it fails to pretty print.
func (token GhToken) PrintTokenAttributes() {
	if err := token.prettyPrintToken(); err != nil {
		log.Printf("error: %v", err)
	}
}

// Print json of GHToken attributes and test validity of token.
func (token GhToken) prettyPrintToken() error {
	token.ValidateSchema()
	if err := fileutil.WriteJSON(os.Stdout, token); err != nil {
		return fmt.Errorf("failed printing token '%s': %w", token.FullToken, err)
	}

	return nil
}
