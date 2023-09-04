// Package cmds_test provides tests for the cmds package.
package cmds_test

import (
	"testing"

	"github.com/pyqlsa/token-forge/internal/cmds"
	"github.com/stretchr/testify/assert"
)

func TestGenerateTokens(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		prefix        string
		validPrefix   bool
		validChecksum bool
		validSchema   bool
	}{
		{
			prefix: "", validPrefix: true, validSchema: true, validChecksum: true,
		},
		{
			prefix: "ghp", validPrefix: true, validSchema: true, validChecksum: true,
		},
		{
			prefix: "gho", validPrefix: true, validSchema: true, validChecksum: true,
		},
		{
			prefix: "ghu", validPrefix: true, validSchema: true, validChecksum: true,
		},
		{
			prefix: "ghs", validPrefix: true, validSchema: true, validChecksum: true,
		},
		{
			prefix: "a", validPrefix: false, validSchema: false, validChecksum: true,
		},
		{
			prefix: "bb", validPrefix: false, validSchema: false, validChecksum: true,
		},
		{
			prefix: "ccc", validPrefix: false, validSchema: false, validChecksum: true,
		},
		{
			prefix: "dddd", validPrefix: false, validSchema: false, validChecksum: true,
		},
	}
	for _, tc := range testcases {
		tc := tc // magic to capture range variable, otherwise we might not actually test all cases
		t.Run(tc.prefix, func(t *testing.T) {
			t.Parallel()
			numTokens := uint64(10000)
			genToken := cmds.GenGhTokenFunc(tc.prefix)
			for i := uint64(0); i < numTokens; i++ {
				fake := genToken()
				assert.Equal(t, fake.HasValidChecksum(), tc.validChecksum, "generated token with invalid checksum: %s", fake.FullToken)
				assert.Equal(t, fake.HasValidPrefix(), tc.validPrefix, "generated token with unexpected prefix validity: %s", fake.FullToken)
				assert.True(t, (len(tc.prefix) < 1 || fake.Prefix == tc.prefix), "generated token with prefix other than requested: %s", fake.FullToken)
				assert.True(t, fake.SchemaChecked, "expected the token's schema to be checked: %s", fake.FullToken)
				assert.Equal(t, fake.SchemaValid, tc.validSchema, "generated token with unexpected schema validity: %s", fake.FullToken)
			}
		})
	}
}
