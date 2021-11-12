// Copyright 2020 FairwindsOps Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	repo := NewRepoList(urls)
	assert.Greater(t, len(repo[0].Charts.Entries), 0)
}

func TestRepo_NewestVersion(t *testing.T) {
	urls := []string{
		"https://charts.fairwinds.com/stable",
	}
	repo := NewRepoList(urls)
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
