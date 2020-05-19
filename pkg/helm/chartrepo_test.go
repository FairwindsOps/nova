package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidRelease(t *testing.T) {
	assert.Equal(t, IsValidRelease("v1.0"), true)
	assert.Equal(t, IsValidRelease("1.0-rc3"), false)
}

func TestNewRepo(t *testing.T) {
	urls := []string{
		"https://charts.fairwinds.com/stable",
	}
	repo := NewRepo(urls)
	assert.Greater(t, len(repo[0].Charts.Entries), 0)
}

func TestRepo_NewestVersion(t *testing.T) {
	urls := []string{
		"https://charts.fairwinds.com/stable",
	}
	repo := NewRepo(urls)
	assert.NotNil(t, repo[0].NewestVersion("rbac-manager"))
}

func TestGetNewestRelease(t *testing.T) {
	repo := Repo{
		Charts: &ChartReleases{
			Entries: map[string][]ChartRelease{
				"foo": {
					{Version: "1.0"},
					{Version: "2.0"},
				},
			},
		},
	}

	newest := repo.NewestVersion("foo")
	assert.Equal(t, "2.0", newest.Version)
}
