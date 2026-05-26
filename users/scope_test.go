package users

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateScopes_SingleScope(t *testing.T) {
	assert.NoError(t, ValidateScopes([]string{"/data"}))
	assert.NoError(t, ValidateScopes([]string{}))
	assert.NoError(t, ValidateScopes(nil))
}

func TestValidateScopes_ValidMultiple(t *testing.T) {
	assert.NoError(t, ValidateScopes([]string{"/data/projects", "/data/media"}))
	assert.NoError(t, ValidateScopes([]string{"/photos", "/documents", "/music"}))
}

func TestValidateScopes_NestedScopesRejected(t *testing.T) {
	err := ValidateScopes([]string{"/data", "/data/photos"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nested")
}

func TestValidateScopes_DuplicateBasenamesRejected(t *testing.T) {
	err := ValidateScopes([]string{"/teamA/docs", "/teamB/docs"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "same folder name")
}

func TestValidateScopes_IdenticalScopesRejected(t *testing.T) {
	err := ValidateScopes([]string{"/data", "/data"})
	assert.Error(t, err)
}

func TestScopeBaseName(t *testing.T) {
	assert.Equal(t, "projects", ScopeBaseName("/data/projects"))
	assert.Equal(t, "media", ScopeBaseName("media"))
	assert.Equal(t, "photos", ScopeBaseName("/photos"))
}
