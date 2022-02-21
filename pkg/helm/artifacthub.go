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
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"k8s.io/klog/v2"
)

const (
	artifactHubAPIRoot         = "https://artifacthub.io"
	maxArtifactHubRequestLimit = 60
	artifactHubHelmKind        = "0"
)

// ArtifactHubPackageClient provides the various pieces to interact with the ArtifactHub API.
type ArtifactHubPackageClient struct {
	APIRoot   string
	URL       *url.URL
	Client    *http.Client
	UserAgent string
}

// ArtifactHubPackageRepo is a simple struct to show a relationship between a helm package name and its repository.
type ArtifactHubPackageRepo struct {
	PackageName string
	RepoName    string
}

// ArtifactHubPackageReturn is the return type for a specific package.
type ArtifactHubPackageReturn struct {
	Package      ArtifactHubHelmPackage
	err          error
	httpResponse *http.Response
}

// ArtifactHubPackagesSearchReturn is the return type for SearchPackages.
type ArtifactHubPackagesSearchReturn struct {
	Packages     []ArtifactHubPackageSearch `json:"packages,omitempty"`
	err          error
	httpResponse *http.Response
}

// ArtifactHubPackageSearch represents a single search return object from the ArtifactHub `packages/search` API.
type ArtifactHubPackageSearch struct {
	PackageID                      string                           `json:"package_id"`
	Name                           string                           `json:"name"`
	NormalizedName                 string                           `json:"normalized_name"`
	LogoImageID                    string                           `json:"logo_image_id"`
	Stars                          int                              `json:"stars"`
	Description                    string                           `json:"description"`
	Version                        string                           `json:"version"`
	AppVersion                     string                           `json:"app_version"`
	Deprecated                     bool                             `json:"deprecated"`
	Signed                         bool                             `json:"signed"`
	SecurityReportSummary          ArtifactHubSecurityReportSummary `json:"security_report_summary"`
	AllContainersImagesWhitelisted bool                             `json:"all_containers_images_whitelisted"`
	ProductionOrganizationsCount   int                              `json:"production_organizations_count"`
	Ts                             int                              `json:"ts"`
	Repository                     ArtifactHubRepository            `json:"repository"`
}

// ArtifactHubSecurityReportSummary is a child struct of ArtifactHubPackageSearch and contains the security report summary for a given package.
type ArtifactHubSecurityReportSummary struct {
	Low      int `json:"low"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Unknown  int `json:"unknown"`
	Critical int `json:"critical"`
}

// ArtifactHubRepository is a child struct of ArtifactHubPackageSearch represents a helm chart repository as provided by the ArtifactHub API.
type ArtifactHubRepository struct {
	URL                     string `json:"url"`
	Kind                    int    `json:"kind"`
	Name                    string `json:"name"`
	Official                bool   `json:"official"`
	DisplayName             string `json:"display_name"`
	RepositoryID            string `json:"repository_id"`
	ScannerDisabled         bool   `json:"scanner_disabled"`
	OrganizationName        string `json:"organization_name"`
	VerifiedPublisher       bool   `json:"verified_publisher"`
	OrganizationDisplayName string `json:"organization_display_name"`
}

// ArtifactHubHelmPackage represents a helm package (chart) as provided by the ArtifactHub API.
type ArtifactHubHelmPackage struct {
	PackageID                      string                           `json:"package_id"`
	Name                           string                           `json:"name"`
	NormalizedName                 string                           `json:"normalized_name"`
	DisplayName                    string                           `json:"display_name"`
	LogoImageID                    string                           `json:"logo_image_id"`
	Stars                          int                              `json:"stars"`
	Description                    string                           `json:"description"`
	Version                        string                           `json:"version"`
	AppVersion                     string                           `json:"app_version"`
	Deprecated                     bool                             `json:"deprecated"`
	Signed                         bool                             `json:"signed"`
	Official                       bool                             `json:"official"`
	ProductionOrganizationsCount   int                              `json:"production_organizations_count"`
	Ts                             int                              `json:"ts"`
	Repository                     ArtifactHubRepository            `json:"repository"`
	SecurityReportSummary          ArtifactHubSecurityReportSummary `json:"security_report_summary"`
	AllContainersImagesWhitelisted bool                             `json:"all_containers_images_whitelisted"`
	LatestVersion                  string                           `json:"latest_version"`
	HomeURL                        string                           `json:"home_url"`
	AvailableVersions              []AvailableVersion               `json:"available_versions"`
	Maintainers                    []Maintainer                     `json:"maintainers"`
	PreRelease                     bool                             `json:"prerelease"`
	Data                           PackageData                      `json:"data"`
	Links                          []Link                           `json:"links"`
}

// AvailableVersion is a sub struct of ArtifactHubHelmPackage and provides a version that is available for a given helm chart.
type AvailableVersion struct {
	Version                 string `json:"version"`
	ContainsSecurityUpdates bool   `json:"contains_security_updates"`
	PreRelease              bool   `json:"prerelease"`
	Ts                      int    `json:"ts"`
}

// Maintainer is a child struct of ArtifactHubHelmPackage and provides information about maintainers of a helm chart.
type Maintainer struct {
	Name         string `json:"name"`
	MaintainerID string `json:"maintainer_id"`
	Email        string `json:"email"`
}

// PackageData is a child struct of ArtifactHubHelmPackage and provides some metadata for a helm chart
type PackageData struct {
	APIVersion   string       `json:"apiVersion"`
	Type         string       `json:"type"`
	KubeVersion  string       `json:"kubeVersion"`
	Dependencies []Dependency `json:"dependencies"`
}

// Dependency is a child struct of PackageData and provides any helm dependency data for a given chart.
type Dependency struct {
	Name                       string `json:"name"`
	Version                    string `json:"version"`
	Repository                 string `json:"repository"`
	ArtifactHubRespositoryName string `json:"artifacthub_respository_name,omitempty"`
}

// Link is child struct of ArtifactHubHelmPackage
type Link struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

// NewArtifactHubPackageClient returns a new client for the unauthenticated paths of the ArtifactHub API.
func NewArtifactHubPackageClient(version string) (*ArtifactHubPackageClient, error) {
	apiRoot := artifactHubAPIRoot
	u, err := url.ParseRequestURI(apiRoot)
	if err != nil {
		return nil, err
	}
	client := new(http.Client)
	return &ArtifactHubPackageClient{
		APIRoot:   apiRoot,
		URL:       u,
		Client:    client,
		UserAgent: fmt.Sprintf("Fairwinds-Nova/%s ", version),
	}, nil
}

// SearchPackages searches for packages given a search term  against the ArtifactHub API.
func (ac *ArtifactHubPackageClient) SearchPackages(searchTerm string) ([]ArtifactHubPackageSearch, error) {
	firstSet := ac.SearchFirst(searchTerm)
	if firstSet.err != nil {
		return nil, firstSet.err
	}
	countHeader := firstSet.httpResponse.Header.Get("Pagination-Total-Count")
	if countHeader == "" {
		klog.V(2).Infof("No Pagination-Total-Count header found in response when searching for '%s' - attempting to return first results", searchTerm)
		return firstSet.Packages, nil
	}
	totalCount, err := strconv.Atoi(countHeader)
	if err != nil {
		return nil, err
	}
	klog.V(5).Infof("found %d packages matching '%s'", totalCount, searchTerm)

	ret := make([]ArtifactHubPackageSearch, totalCount)
	for i, p := range firstSet.Packages {
		ret[i] = p
	}
	if totalCount > maxArtifactHubRequestLimit {
		wg := sync.WaitGroup{}
		for i := maxArtifactHubRequestLimit; i < totalCount; i += maxArtifactHubRequestLimit {
			wg.Add(1)
			klog.V(8).Infof("paging API for search term '%s' at %d offset", searchTerm, i)
			go func(offset int, sTerm string, wg *sync.WaitGroup, ret *[]ArtifactHubPackageSearch) {
				defer wg.Done()
				searchReturn := ac.Search(sTerm, offset)
				if searchReturn.err != nil {
					klog.V(3).Infof("error searching for term '%s': %s", sTerm, searchReturn.err)
				}
				for i, p := range searchReturn.Packages {
					if offset+i > len(*ret) {
						continue
					}
					(*ret)[offset+i] = p
				}
			}(i, searchTerm, &wg, &ret)
		}
		wg.Wait()
	}
	return ret, nil
}

// SearchForPackageRepo calls SearchPackages with a given searchTerm, and then filters the results to only return the package name and repository info.
func (ac *ArtifactHubPackageClient) SearchForPackageRepo(searchTerm string) ([]ArtifactHubPackageRepo, error) {
	packages, err := ac.SearchPackages(searchTerm)
	if err != nil {
		return nil, err
	}
	ret := make([]ArtifactHubPackageRepo, 0)
	for _, p := range packages {
		r := ArtifactHubPackageRepo{
			PackageName: p.Name,
			RepoName:    p.Repository.Name,
		}
		ret = append(ret, r)
	}
	return ret, nil
}

// MultiSearch will find all packages that match various search terms (terms are not combined, but searched individually).
// Returns only the package name and repository information.
func (ac *ArtifactHubPackageClient) MultiSearch(searchTerms []string) ([]ArtifactHubPackageRepo, error) {
	ret := make([]ArtifactHubPackageRepo, 0)
	termMap := make(map[string][]ArtifactHubPackageRepo)
	wg := sync.WaitGroup{}
	klog.V(10).Infof("filtering MultiSearch string array duplicates. starting amount: %d", len(searchTerms))
	searchTerms = removeDuplicateString(searchTerms)
	klog.V(10).Infof("filtered down to %d search terms", len(searchTerms))
	mut := sync.Mutex{}
	for _, searchTerm := range searchTerms {
		wg.Add(1)
		go func(wg *sync.WaitGroup, term string, r *map[string][]ArtifactHubPackageRepo, m *sync.Mutex) {
			defer wg.Done()
			packages, err := ac.SearchForPackageRepo(term)
			if err != nil {
				klog.V(3).Infof("error searching for term %s", err)
				(*r)[term] = nil
				return
			}
			if _, exists := (*r)[term]; !exists {
				m.Lock()
				defer m.Unlock()
				(*r)[term] = packages
			}

		}(&wg, searchTerm, &termMap, &mut)
	}
	wg.Wait()
	for term, packageRepos := range termMap {
		if packageRepos == nil {
			return nil, fmt.Errorf("failed to search for packages for term %s", term)
		}
		ret = append(ret, packageRepos...)
	}
	return ret, nil
}

// SearchFirst will find the first set (page) of packages that match a search term.
func (ac *ArtifactHubPackageClient) SearchFirst(searchTerm string) ArtifactHubPackagesSearchReturn {
	return ac.Search(searchTerm, 0)
}

// Search makes use of the package search API: https://artifacthub.io/docs/api/#/Packages/searchPackages
// It sets up the proper query parameters and adds the search tearm to the ts_query_web parameter.
func (ac *ArtifactHubPackageClient) Search(searchTerm string, offset int) (ret ArtifactHubPackagesSearchReturn) {
	urlValues := url.Values{}
	urlValues.Add("ts_query_web", searchTerm)
	urlValues.Add("kind", artifactHubHelmKind)
	urlValues.Add("limit", strconv.Itoa(maxArtifactHubRequestLimit))
	urlValues.Add("offset", strconv.Itoa(offset))
	urlValues.Add("sort", "stars")
	searchPath := "api/v1/packages/search"
	resp, err := ac.get(searchPath, urlValues)
	if err != nil {
		klog.V(3).Infof("error GETing response for with search term '%s': %s", searchTerm, err)
		return ArtifactHubPackagesSearchReturn{
			err:          err,
			httpResponse: nil,
		}
	}
	if resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		err := json.NewDecoder(resp.Body).Decode(&ret)
		if err != nil {
			klog.V(3).Infof("error decoding search json with search term '%s': %s", searchTerm, err)
			return ArtifactHubPackagesSearchReturn{
				err:          err,
				httpResponse: resp,
			}
		}
		ret.httpResponse = resp
		return
	}
	klog.V(3).Infof("got non-200 response searching for term '%s': %d", searchTerm, resp.StatusCode)
	return ArtifactHubPackagesSearchReturn{
		err:          fmt.Errorf("error code: %d", resp.StatusCode),
		httpResponse: resp,
	}
}

// GetPackages makes use of the helm package details API: https://artifacthub.io/docs/api/#/Packages/getHelmPackageDetails
// It sets up the proper query parameters and adds the repo/package name to the path.
func (ac *ArtifactHubPackageClient) GetPackages(packageRepos []ArtifactHubPackageRepo) []ArtifactHubHelmPackage {
	ret := make([]ArtifactHubHelmPackage, len(packageRepos))
	wg := sync.WaitGroup{}
	for i, r := range packageRepos {
		wg.Add(1)
		go func(index int, repo ArtifactHubPackageRepo, wg *sync.WaitGroup, ret *[]ArtifactHubHelmPackage) {
			defer wg.Done()
			response := ac.getSpecific(fmt.Sprintf("api/v1/packages/helm/%s/%s", repo.RepoName, repo.PackageName))
			if response.err != nil {
				klog.V(3).Infof("error getting package %s/%s: %s", repo.RepoName, repo.PackageName, response.err)
			}
			(*ret)[index] = response.Package
		}(i, r, &wg, &ret)
	}
	wg.Wait()
	return ret
}

func (ac *ArtifactHubPackageClient) getSpecific(path string) (ret ArtifactHubPackageReturn) {
	klog.V(10).Infof("getting package %s", path)
	resp, err := ac.get(path, nil)
	if err != nil {
		klog.V(3).Infof("error GETing response for path '%s': %s", path, err)
		return ArtifactHubPackageReturn{
			err:          err,
			httpResponse: nil,
		}
	}
	if resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		err := json.NewDecoder(resp.Body).Decode(&ret.Package)
		if err != nil {
			klog.V(3).Infof("error decoding response for path %s:\n%v", path, err)
			return ArtifactHubPackageReturn{
				err:          err,
				httpResponse: resp,
			}
		}
		ret.httpResponse = resp
		return
	}
	klog.V(3).Infof("got non-200 response for path %s: %d", path, resp.StatusCode)
	return ArtifactHubPackageReturn{
		err:          fmt.Errorf("error code: %d", resp.StatusCode),
		httpResponse: resp,
	}
}

// get is the basic getter for the artifacthub package client
// The path argument should be formatted like so: "api/v1/packages/search", any unauthenticated paths
// will work and are documented here: https://artifacthub.io/docs/api/#/
// urlValues are the search parameters for the query if necessary.
// offset is to be used for pagination. The first page would be offset 0.
func (ac *ArtifactHubPackageClient) get(path string, urlValues url.Values) (*http.Response, error) {
	requestURL := *ac.URL
	requestURL.Path = path
	urlString := requestURL.String()
	r, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		return nil, err
	}
	q := r.URL.Query()
	for k, v := range urlValues {
		for _, vv := range v {
			q.Add(k, vv)
		}
	}
	r.URL.RawQuery = q.Encode()
	r.Header.Add("accept", "application/json")
	r.Header.Set("User-Agent", ac.UserAgent)
	var response *http.Response
	for attempt := 1; attempt <= 5; attempt++ {
		resp, innerErr := ac.Client.Do(r)
		if innerErr != nil {
			response = nil
			err = innerErr
			klog.V(3).Infof("attempt %d failed to GET %s: %v", attempt, urlString, err)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			response = resp
			klog.V(3).Infof("attempt %d failed to GET %s with status code: %v", attempt, urlString, resp.StatusCode)
			err = fmt.Errorf("error code: %d", resp.StatusCode)
			continue
		}
		response = resp
		break
	}
	return response, err
}

func removeDuplicateString(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
