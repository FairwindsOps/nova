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

package containers

import (
	"context"
	"reflect"
	"testing"

	version "github.com/Masterminds/semver/v3"
	"github.com/fairwindsops/nova/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	testNamespace          = "test"
	testPodName            = "test-pod"
	testInitContainerName  = "test-init-container"
	testInitContainerImage = "test-init-container-image:v1.0.0"
	testContainerName      = "test-container"
	testContainerImage     = "test-image:v1.0.0"
)

var (
	testClient = &Client{
		Kube: kube.SetAndGetMock(),
	}
	testPodSpec = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testPodName,
			Namespace: testNamespace,
		},
		Spec: corev1.PodSpec{
			InitContainers: []corev1.Container{
				{
					Name:  testInitContainerName,
					Image: testInitContainerImage,
				},
			},
			Containers: []corev1.Container{
				{
					Name:  testContainerName,
					Image: testContainerImage,
				},
			},
		},
	}
	testImageStruct = Image{
		Name:   "test-image",
		Prefix: "v",
		Current: &Tag{
			version: version.MustParse("1.0.0"),
			Value:   "1.0.0",
		},
		allTags: []string{
			"v1.0.0",
			"v1.0.1",
			"v1.0.2",
			"v1.1.0",
			"v2.0.0",
		},
	}
	testImageStructNonSemver = Image{
		Name:   "test-image",
		Prefix: "v",
		Current: &Tag{
			version: version.MustParse("1.0.0"),
			Value:   "1.0.0",
		},
		allTags: []string{
			"test-1.0.0",
			"bad",
			"v1.0-debian-1",
			"v1.1.0",
			"v2.0.0",
		},
	}
)

func TestGetContainerImages(t *testing.T) {
	setupKubeObjects(t, testClient)
	defer teardownKubeObjects(t, testClient)

	tests := []struct {
		name    string
		want    []string
		wantErr bool
	}{
		{
			name:    "TestGetContainerImages",
			want:    []string{testInitContainerImage, testContainerImage},
			wantErr: bool(false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := testClient.getContainerImages()
			if (err != nil) != tt.wantErr {
				t.Errorf("getContainerImages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getContainerImages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveDuplicateStr(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "TestRemoveDuplicateString",
			input: []string{"a", "b", "c", "a"},
			want:  []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := removeDuplicateStr(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("removeDuplicateString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewImage(t *testing.T) {
	tests := []struct {
		name         string
		fullImageTag string
		want         *Image
		wantErr      bool
	}{
		{
			name:         "TestNewImage_Good",
			fullImageTag: testContainerImage,
			want: &Image{
				Name:   "test-image",
				Prefix: "v",
				Current: &Tag{
					Value: "1.0.0",
				},
				StrictSemver: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newImage(tt.fullImageTag)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Name != tt.want.Name {
				t.Errorf("NewImage() Name got = %v, want %v", got.Name, tt.want.Name)
			}
			if got.Prefix != tt.want.Prefix {
				t.Errorf("NewImage() Prefix got = %v, want %v", got.Prefix, tt.want.Prefix)
			}
			if got.Current.Value != tt.want.Current.Value {
				t.Errorf("NewImage() Current.Value got = %v, want %v", got.Current.Value, tt.want.Current.Value)
			}
			if got.StrictSemver != tt.want.StrictSemver {
				t.Errorf("NewImage() StrictSemver got = %v, want %v", got.StrictSemver, tt.want.StrictSemver)
			}
		})
	}
}

func TestParseTags(t *testing.T) {
	// Copying these to make sure we don't inadvertently modify the top level vars
	testImageStruct := testImageStruct
	testImageStructNonSemver := testImageStructNonSemver
	// Reference the newly copied vars location in memory
	testImageStructPointer := &testImageStruct
	testImageStructNonSemverPointer := &testImageStructNonSemver
	tests := []struct {
		name                string
		image               *Image
		wantSemverTagLen    int
		wantNonSemverTagLen int
	}{
		{
			name:                "TestParseTags",
			image:               testImageStructPointer,
			wantSemverTagLen:    5,
			wantNonSemverTagLen: 0,
		},
		{
			name:                "TestParseTagsNonSemver",
			image:               testImageStructNonSemverPointer,
			wantSemverTagLen:    3,
			wantNonSemverTagLen: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.image.parseTags()
			got := tt.image
			if len(got.semverTags) != tt.wantSemverTagLen {
				t.Errorf("ParseTags() SemverTags got = %v, want %v, values got = %v", len(got.semverTags), tt.wantSemverTagLen, got.semverTags)
			}
			if len(got.nonSemverTags) != tt.wantNonSemverTagLen {
				t.Errorf("ParseTags() NonSemverTags got = %v, want %v, values got = %v", len(got.nonSemverTags), tt.wantNonSemverTagLen, got.nonSemverTags)
			}
		})
	}
}

func TestPopulateNewest(t *testing.T) {
	// Copying this to make sure we don't inadvertently modify the top level vars
	testImageStruct := testImageStruct
	// Reference the newly copied var location in memory
	testImageStructPointer := &testImageStruct
	tests := []struct {
		name        string
		image       *Image
		newestTag   string
		newestMinor string
		newestPatch string
		wantErr     bool
	}{
		{
			name:        "TestPopulateNewest",
			image:       testImageStructPointer,
			newestTag:   "2.0.0",
			newestMinor: "1.1.0",
			newestPatch: "1.0.2",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.image.parseTags()
			err := tt.image.populateNewest()
			got := tt.image
			if (err != nil) != tt.wantErr {
				t.Errorf("populateNewest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Newest.Value != tt.newestTag {
				t.Errorf("populateNewest() Newest.Value got = %v, want %v", got.Newest.Value, tt.newestTag)
			}
			if got.NewestMinor.Value != tt.newestMinor {
				t.Errorf("populateNewest() NewestMinor.Value got = %v, want %v", got.NewestMinor.Value, tt.newestMinor)
			}
			if got.NewestPatch.Value != tt.newestPatch {
				t.Errorf("populateNewest() NewestPatch.Value got = %v, want %v", got.NewestPatch.Value, tt.newestPatch)
			}
		})
	}
}

func TestParseTagString(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantVerString string
		wantBool      bool
		wantErr       bool
	}{
		{
			name:          "StrictTrue",
			input:         "1.0.0",
			wantVerString: "1.0.0",
			wantBool:      bool(true),
		},
		{
			name:          "StrictFalse",
			input:         "v1.0.0",
			wantVerString: "v1.0.0",
			wantBool:      bool(false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotString, gotBool := parseTagString(tt.input)
			if gotString != tt.wantVerString || gotBool != tt.wantBool {
				t.Errorf("parseTagString() = %v and %v, want %v and %v", gotString, gotBool, tt.wantVerString, tt.wantBool)
			}
		})
	}
}

func TestPreReleaseRegex(t *testing.T) {
	tests := []struct {
		name             string
		preReleaseIgnore []string
		inputPreRelease  string
		want             bool
	}{
		{
			name:             "TestPreReleaseRegex",
			preReleaseIgnore: []string{"alpha", "beta", "rc"},
			inputPreRelease:  "alpha",
			want:             bool(true),
		},
		{
			name:             "TestPreReleaseRegexFalse",
			preReleaseIgnore: []string{"foobar", "alphax"},
			inputPreRelease:  "alpha",
			want:             bool(false),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := preReleaseRegex(tt.preReleaseIgnore, tt.inputPreRelease); got != tt.want {
				t.Errorf("preReleaseRegex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func setupKubeObjects(t *testing.T, c *Client) {
	k := c.Kube.Client.(*fake.Clientset)
	_, err := k.CoreV1().Pods(testNamespace).Create(context.TODO(), testPodSpec, metav1.CreateOptions{})
	if err != nil {
		t.Errorf("Error creating pod: %v", err)
	}
}

func teardownKubeObjects(t *testing.T, c *Client) {
	k := c.Kube.Client.(*fake.Clientset)
	err := k.CoreV1().Pods(testNamespace).Delete(context.TODO(), testPodName, metav1.DeleteOptions{})
	if err != nil {
		t.Errorf("Error deleting pod: %v", err)
	}
}
