package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHubConfig(t *testing.T) {
	_, iErr := NewHubConfig("invalid-url")
	assert.NotNil(t, iErr)

	valid, err := NewHubConfig("https://raw.githubusercontent.com/helm/hub/master/config/repo-values.yaml")
	assert.Nil(t, err)
	assert.Greater(t, len(valid.Sync.Repos), 0)
}
