package containers

import (
	"context"
	"reflect"
	"testing"

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
