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

	assert.Equal(t, 12600, len(resp))
	toCheck := resp[6467]
	assert.Equal(t, "open5gs-webui", toCheck.Name)
	assert.Equal(t, "Helm chart to deploy Open5gs WebUI service on Kubernetes. ", toCheck.Description)

	assert.Equal(t, 1, len(toCheck.Links))
	assert.Equal(t, "http://open5gs.org", toCheck.Links[0].URL)
	assert.Equal(t, "source", toCheck.Links[0].Name)

	assert.Equal(t, 2, len(toCheck.Maintainers))
	assert.Equal(t, "cgiraldo", toCheck.Maintainers[0].Name)

	assert.Equal(t, "open5gs-webui", toCheck.Repository.Name)
	assert.Equal(t, "oci://registry-1.docker.io/gradiant/open5gs-webui", toCheck.Repository.URL)

	assert.Equal(t, "2.2.0", toCheck.Version)
	assert.Equal(t, "2.7.0", toCheck.AppVersion)
	assert.Equal(t, 7, len(toCheck.AvailableVersions))
	assert.Equal(t, "2.0.0", toCheck.AvailableVersions[0].Version)

	toCheckWithKubeVersion := resp[10]
	assert.Equal(t, "ndb-operator", toCheckWithKubeVersion.Name)
	assert.Equal(t, "1.3.0", toCheckWithKubeVersion.Version)
	assert.Equal(t, "8.3.0-1.3.0", toCheckWithKubeVersion.AppVersion)
	assert.Equal(t, ">= 1.23.0-0", toCheckWithKubeVersion.KubeVersion)
	assert.Len(t, toCheckWithKubeVersion.AvailableVersions, 2)
}
