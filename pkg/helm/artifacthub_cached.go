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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"k8s.io/klog/v2"
)

const (
	artifactHubCachedAPIRoot         = "https://artifacthub.io/api/v1/nova"
	maxArtifactHubCachedRequestLimit = 60
	artifactHubCachedHelmKind        = "0"
)

var cacheFile = ""

func init() {
	if os.Getenv("ARTIFACT_HUB_CACHE_FILE") != "" {
		cacheFile = os.Getenv("ARTIFACT_HUB_CACHE_FILE")
	}
}

// ArtifactHubCachedPackageClient provides the various pieces to interact with the ArtifactHubCached API.
type ArtifactHubCachedPackageClient struct {
	APIRoot   string
	URL       *url.URL
	Client    *http.Client
	UserAgent string
}

// ArtifactHubCachedPackagesList contains the output from the AH cached API
type ArtifactHubCachedPackagesList []ArtifactHubCachedPackage

// ArtifactHubCachedPackage represents a single entry in the API output. It's a single chart registered in AH
type ArtifactHubCachedPackage struct {
	Name          string                         `json:"name"`
	Description   string                         `json:"description"`
	HomeURL       string                         `json:"home"`
	Repository    ArtifactHubCachedRepository    `json:"repository"`
	Official      bool                           `json:"official"`
	LatestVersion string                         `json:"latest_version"`
	Versions      []ArtifactHubCachedVersionInfo `json:"versions"`
	Links         []Link                         `json:"links"`
	Maintainers   []Maintainer                   `json:"maintainers"`
	Deprecated    bool                           `json:"deprecated"`
}

// ArtifactHubCachedRepository is a sub-struct of the Package struct, and represents the repository containing the package.
type ArtifactHubCachedRepository struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Official bool   `json:"official"`
	Verified bool   `json:"verified"`
}

// ArtifactHubCachedVersionInfo represents the chart and application version of a package
type ArtifactHubCachedVersionInfo struct {
	Version    string `json:"pkg"`
	AppVersion string `json:"app"`
	Deprecated bool   `json:"deprecated"`
}

// NewArtifactHubCachedPackageClient returns a new client for the unauthenticated paths of the ArtifactHubCached API.
func NewArtifactHubCachedPackageClient(version string) (*ArtifactHubCachedPackageClient, error) {
	apiRoot := artifactHubCachedAPIRoot
	u, err := url.ParseRequestURI(apiRoot)
	if err != nil {
		return nil, err
	}
	client := new(http.Client)
	return &ArtifactHubCachedPackageClient{
		APIRoot:   apiRoot,
		URL:       u,
		Client:    client,
		UserAgent: fmt.Sprintf("Fairwinds-Nova/%s ", version),
	}, nil
}

// List returns all packages from ArtifactHub
func (ac *ArtifactHubCachedPackageClient) List() ([]ArtifactHubHelmPackage, error) {
	list := ArtifactHubCachedPackagesList{}
	if cacheFile == "" {
		resp, err := ac.get()
		if err != nil {
			return nil, err
		}
		err = json.NewDecoder(resp.Body).Decode(&list)
		if err != nil {
			return nil, err
		}
	} else {
		cache, err := ioutil.ReadFile(cacheFile)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(cache, &list)
		if err != nil {
			return nil, err
		}
	}
	packages := make([]ArtifactHubHelmPackage, len(list))
	for idx, cachedPackage := range list {
		packages[idx] = ArtifactHubHelmPackage{
			Name:        cachedPackage.Name,
			Description: cachedPackage.Description,
			Maintainers: cachedPackage.Maintainers,
			HomeURL:     cachedPackage.HomeURL,
			Links:       cachedPackage.Links,
			Official:    cachedPackage.Official,
			Repository: ArtifactHubRepository{
				Name:              cachedPackage.Repository.Name,
				URL:               cachedPackage.Repository.URL,
				VerifiedPublisher: cachedPackage.Repository.Verified,
				Official:          cachedPackage.Repository.Official,
			},
			Version:           cachedPackage.LatestVersion,
			AvailableVersions: []AvailableVersion{},
		}
		for _, version := range cachedPackage.Versions {
			if version.Version == cachedPackage.LatestVersion {
				packages[idx].AppVersion = version.AppVersion
				packages[idx].Deprecated = version.Deprecated
			}
			packages[idx].AvailableVersions = append(packages[idx].AvailableVersions, AvailableVersion{
				Version: version.Version,
			})
		}
	}
	return packages, nil
}

// get is the basic getter for the artifacthub cached package client
func (ac *ArtifactHubCachedPackageClient) get() (*http.Response, error) {
	requestURL := *ac.URL
	urlString := requestURL.String()
	r, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		return nil, err
	}
	q := r.URL.Query()
	r.URL.RawQuery = q.Encode()
	r.Header.Add("accept", "application/json")
	r.Header.Set("User-Agent", ac.UserAgent)
	resp, err := ac.Client.Do(r)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		klog.V(3).Infof("failed to GET %s with status code: %v", urlString, resp.StatusCode)
		err = fmt.Errorf("error code: %d", resp.StatusCode)
	}
	return resp, err
}
