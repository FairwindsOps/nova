package helm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"k8s.io/klog"
)

const (
	artifactHubAPIRoot         = "https://artifacthub.io"
	maxArtifactHubRequestLimit = 60
	artifactHubHelmKind        = "0"
)

var t = time.Now()

type ArtifactHubPackageClient struct {
	APIRoot string
	URL     *url.URL
	Client  *http.Client
}

type ArtifactHubPackageRepo struct {
	PackageName string
	RepoName    string
}

type ArtifactHubPackageReturn struct {
	Package      ArtifactHubHelmPackage
	err          error
	httpResponse *http.Response
}

type ArtifactHubPackagesSearchReturn struct {
	Packages     []ArtifactHubPackageSearch `json:"packages,omitempty"`
	err          error
	httpResponse *http.Response
}

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

type ArtifactHubSecurityReportSummary struct {
	Low      int `json:"low"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Unknown  int `json:"unknown"`
	Critical int `json:"critical"`
}

type ArtifactHubRepository struct {
	Url                     string `json:"url"`
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

type AvailableVersion struct {
	Version                 string `json:"version"`
	ContainsSecurityUpdates bool   `json:"contains_security_updates"`
	PreRelease              bool   `json:"prerelease"`
	Ts                      int    `json:"ts"`
}

type Maintainer struct {
	Name         string `json:"name"`
	MaintainerID string `json:"maintainer_id"`
	Email        string `json:"email"`
}

type PackageData struct {
	APIVersion   string       `json:"apiVersion"`
	Type         string       `json:"type"`
	KubeVersion  string       `json:"kubeVersion"`
	Dependencies []Dependency `json:"dependencies"`
}

type Dependency struct {
	Name                       string `json:"name"`
	Version                    string `json:"version"`
	Repository                 string `json:"repository"`
	ArtifactHubRespositoryName string `json:"artifacthub_respository_name,omitempty"`
}

type Link struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

// NewArtifactHubPackageClient returns a new client for the unauthenticated paths of the ArtifactHub API.
func NewArtifactHubPackageClient() (*ArtifactHubPackageClient, error) {
	apiRoot := artifactHubAPIRoot
	u, err := url.ParseRequestURI(apiRoot)
	if err != nil {
		return nil, err
	}
	client := new(http.Client)
	return &ArtifactHubPackageClient{
		APIRoot: apiRoot,
		URL:     u,
		Client:  client,
	}, nil
}

func (ac *ArtifactHubPackageClient) SearchPackages(searchTerm string) ([]ArtifactHubPackageSearch, error) {
	firstSet := ac.SearchFirst(searchTerm)
	if firstSet.err != nil {
		return nil, firstSet.err
	}
	countHeader := firstSet.httpResponse.Header.Get("Pagination-Total-Count")
	if countHeader == "" {
		return []ArtifactHubPackageSearch{}, nil
	}
	totalCount, err := strconv.Atoi(countHeader)
	if err != nil {
		return nil, err
	}

	ret := make([]ArtifactHubPackageSearch, totalCount)
	for i, p := range firstSet.Packages {
		ret[i] = p
	}
	if totalCount > maxArtifactHubRequestLimit {
		wg := sync.WaitGroup{}
		for i := maxArtifactHubRequestLimit; i < totalCount; i += maxArtifactHubRequestLimit {
			wg.Add(1)
			go func(offset int, wg *sync.WaitGroup, ret *[]ArtifactHubPackageSearch) {
				defer wg.Done()
				for i, p := range ac.Search(searchTerm, offset).Packages {
					if offset+i > len(*ret) {
						continue
					}
					(*ret)[offset+i] = p
				}
			}(i, &wg, &ret)
		}
		wg.Wait()
	}
	return ret, nil
}

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

func (ac *ArtifactHubPackageClient) MultiSearch(searchTerms []string) ([]ArtifactHubPackageRepo, error) {
	ret := make([]ArtifactHubPackageRepo, 0)
	termMap := make(map[string][]ArtifactHubPackageRepo)
	wg := sync.WaitGroup{}
	for _, searchTerm := range searchTerms {
		wg.Add(1)
		go func(wg *sync.WaitGroup, term string, r *map[string][]ArtifactHubPackageRepo) {
			defer wg.Done()
			packages, err := ac.SearchForPackageRepo(term)
			klog.V(3).Infof("found %d packages searching for term %s", len(packages), term)
			if err != nil {
				klog.Errorf("error searching for term %s", err)
				(*r)[term] = nil
				return
			}
			(*r)[term] = packages
		}(&wg, searchTerm, &termMap)
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
		return ArtifactHubPackagesSearchReturn{
			err:          err,
			httpResponse: nil,
		}
	}
	if resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		err := json.NewDecoder(resp.Body).Decode(&ret)
		if err != nil {
			return ArtifactHubPackagesSearchReturn{
				err:          err,
				httpResponse: resp,
			}
		}
		ret.httpResponse = resp
		return
	} else {
		return ArtifactHubPackagesSearchReturn{
			err:          fmt.Errorf("error code: %d", resp.StatusCode),
			httpResponse: resp,
		}
	}
}

func (ac *ArtifactHubPackageClient) GetPackages(packageRepos []ArtifactHubPackageRepo) []ArtifactHubHelmPackage {
	ret := make([]ArtifactHubHelmPackage, len(packageRepos))
	wg := sync.WaitGroup{}
	for i, r := range packageRepos {
		wg.Add(1)
		go func(index int, repo ArtifactHubPackageRepo, wg *sync.WaitGroup, ret *[]ArtifactHubHelmPackage) {
			defer wg.Done()
			response := ac.getSpecific(fmt.Sprintf("api/v1/packages/helm/%s/%s", repo.RepoName, repo.PackageName))
			if response.err != nil {
				klog.Errorf("error getting package %s/%s: %s", repo.RepoName, repo.PackageName, response.err)
			}
			(*ret)[index] = response.Package
		}(i, r, &wg, &ret)
	}
	wg.Wait()
	return ret
}

func (ac *ArtifactHubPackageClient) getSpecific(path string) (ret ArtifactHubPackageReturn) {
	klog.V(8).Infof("getting package %s", path)
	resp, err := ac.get(path, nil)
	if err != nil {
		return ArtifactHubPackageReturn{
			err:          err,
			httpResponse: nil,
		}
	}
	if resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		err := json.NewDecoder(resp.Body).Decode(&ret.Package)
		if err != nil {
			klog.Errorf("error decoding response for path %s:\n%v", path, err)
			return ArtifactHubPackageReturn{
				err:          err,
				httpResponse: resp,
			}
		}
		ret.httpResponse = resp
		if path == "api/v1/packages/helm/bitnami/metrics-server" || path == "api/v1/packages/helm/bitnami/redis" {
			if ret.Package.Repository.Name != "bitnami" {
				klog.Warningf("GET %s %v Repo: %s Request ID: %s", path, ret.httpResponse.StatusCode, ret.Package.Repository.Name)
			}
		}
		return
	} else {
		klog.Errorf("error GETing response for path %s: %d", path, resp.StatusCode)
		return ArtifactHubPackageReturn{
			err:          fmt.Errorf("error code: %d", resp.StatusCode),
			httpResponse: resp,
		}
	}
}

// get is the basic getter for the artifacthub package client
// The path argument should be formatted like so: "api/v1/packages/search", any unauthenticated paths
// will work and are documented here: https://artifacthub.io/docs/api/#/
// urlValues are the search parameters for the query if necessary.
// offset is to be used for pagination. The first page would be offset 0.
func (ac *ArtifactHubPackageClient) get(path string, urlValues url.Values) (*http.Response, error) {
	requestUrl := *ac.URL
	requestUrl.Path = path
	urlString := requestUrl.String()
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

func (p ArtifactHubHelmPackage) VersionExists(version string) bool {
	for _, v := range p.AvailableVersions {
		if v.Version == version {
			return true
		}
	}
	return false
}
