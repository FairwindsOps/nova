package output

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileOutput_Send(t *testing.T) {
	path := "/tmp/output.json"

	out := Output{
		HelmReleases: []ReleaseOutput{
			{
				ReleaseName: "foo",
				Namespace:   "foo",
				Installed: VersionInfo{
					Version: "1.0",
				},
				Latest: VersionInfo{
					Version: "1.0",
				},
				Home:        "https://wiki.example.com",
				Deprecated:  false,
				Description: "Test description for foo chart",
				Icon:        "https://wiki.example.com/logo.png",
				IsOld:       false,
			},
		},
	}

	err := out.ToFile(path)
	assert.Nil(t, err)

	_, existsErr := os.Stat(path)
	assert.Nil(t, existsErr)

	if existsErr == nil {
		os.Remove(path)
	}

}
