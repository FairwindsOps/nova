package helm

import (
	"io/ioutil"
	"net/http"
	"strings"

	"gopkg.in/yaml.v2"

	"k8s.io/klog"
)

// HubConfig contains the sync config for helm hub
type HubConfig struct {
	Sync SyncConfig `yaml:"sync"`
}

// SyncConfig contains the config used to sync helm hub
type SyncConfig struct {
	Repos []HubSyncedRepository `yaml:"repos"`
}

// HubSyncedRepository contains information about a helm repo that publishes to helm hub
type HubSyncedRepository struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

// URLs returns a collection of helm repos that publish to Helm Hub
func (h *HubConfig) URLs() []string {
	urls := []string{}

	for _, repo := range h.Sync.Repos {
		// add url.  Trim trailing "/" if present.
		urls = append(urls, strings.TrimSuffix(repo.URL, "/"))
	}
	return urls
}

// NewHubConfig returns a helm hub sync config
func NewHubConfig(url string) (*HubConfig, error) {
	cfg := HubConfig{}
	response, err := http.Get(url)
	if err != nil {
		klog.Warningf("Error loading HubConfig sync %s: %v", url, err)
		return nil, err
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		klog.Warningf("Error reading HubConfig data: %v", err)
		return nil, err
	}

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		klog.Warningf("Error unmarshaling yaml for hub sync: %v", err)
		return nil, err
	}
	return &cfg, nil
}
