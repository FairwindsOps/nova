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
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

const sampleFile = "artifacthub_cached_sample.json"

func Test_ingestSample(t *testing.T) {
	sample, err := ioutil.ReadFile(sampleFile)
	if err != nil {
		panic(err)
	}
	resp := ArtifactHubCachedPackagesList{}
	err = json.Unmarshal(sample, &resp)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(resp))
	assert.Equal(t, "secrets-store-csi-driver-provider-gc", resp[0].Name)
}
