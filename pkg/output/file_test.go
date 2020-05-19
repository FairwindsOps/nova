package output

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileOutput_Send(t *testing.T) {
	fArgs := FileArgs{
		Path: "/tmp/output.json",
	}

	fo := FileOutput{
		Output: []ReleaseOutput{
			{
				ReleaseName:   "foo",
				Namespace:     "foo",
				Version:       "1.0",
				Home:          "https://wiki.example.com",
				Deprecated:    false,
				Description:   "Test description for foo chart",
				Icon:          "https://wiki.example.com/logo.png",
				NewestVersion: "1.0",
				IsOld:         false,
			},
		},
	}

	err := fo.Send(&fArgs)
	assert.Nil(t, err)

	_, existsErr := os.Stat(fArgs.Path)
	assert.Nil(t, existsErr)

	if existsErr == nil {
		os.Remove(fArgs.Path)
	}

}
