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

const sampleFile = "artifacthub_cached_sample.json"

func Test_ingestSample(t *testing.T) {
	cacheFile = sampleFile
	client, err := NewArtifactHubCachedPackageClient("")
	assert.NoError(t, err)
	resp, err := client.List()
	assert.NoError(t, err)

	assert.Equal(t, 11233, len(resp))
	toCheck := resp[1]
	assert.Equal(t, "open5gs-webui", toCheck.Name)
	assert.Equal(t, "Helm chart to deploy Open5gs WebUI service on Kubernetes. ", toCheck.Description)

	assert.Equal(t, 1, len(toCheck.Links))
	assert.Equal(t, "http://open5gs.org", toCheck.Links[0].URL)
	assert.Equal(t, "source", toCheck.Links[0].Name)

	assert.Equal(t, 2, len(toCheck.Maintainers))
	assert.Equal(t, "cgiraldo", toCheck.Maintainers[0].Name)

	assert.Equal(t, "gradiant-openverso", toCheck.Repository.Name)
	assert.Equal(t, "https://gradiant.github.io/openverso-charts/", toCheck.Repository.URL)

	assert.Equal(t, "2.0.3", toCheck.Version)
	assert.Equal(t, "2.4.11", toCheck.AppVersion)
	assert.Equal(t, 4, len(toCheck.AvailableVersions))
	assert.Equal(t, "2.0.0", toCheck.AvailableVersions[0].Version)
}
