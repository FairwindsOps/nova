package containers

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	version "github.com/Masterminds/semver/v3"
	"github.com/fairwindsops/nova/pkg/kube"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

var preReleaseIgnore = []string{"alpha", "beta", "rc", "snapshot", "dev", "prerelease", "pre"}

// Client represents a kubernetes client. Having a struct around this allows us to implement a fake client in tests
type Client struct {
	Kube *kube.Connection
}

// Results is a struct that contains a list of Images and a list of ErroredImages. This is the main thing that is returned from this package
type Results struct {
	Images    []*Image
	ErrImages []*ErroredImage
}

// ErroredImage is an image that failed to parse for any number of reasons. The error message is captured for later logging
type ErroredImage struct {
	Image string
	Err   string
}

// Image contains all the relevant data for reporting an out of date image
type Image struct {
	Name          string
	Prefix        string
	Current       *Tag
	Newest        *Tag
	NewestPatch   *Tag
	NewestMinor   *Tag
	StrictSemver  bool
	semverTags    []*version.Version
	nonSemverTags []string
	repo          name.Repository
	allTags       []string
}

// PodData represents a pod and it's images so that we can report the namespace and other information later
type PodData struct {
	Name           string
	Namespace      string
	InitContainers []string
	Containers     []string
}

// Tag represents one single tag of a container image
type Tag struct {
	version *version.Version
	Value   string
}

// NewClient is a constructor to create a new Client
func NewClient(kubeContext string) *Client {
	return &Client{
		Kube: kube.GetConfigInstance(kubeContext),
	}
}

// Find is the primary function for this package that returns the results of images found in the cluster and whether they are out of date or not
func (c *Client) Find() (Results, error) {
	clusterImages, err := c.getContainerImages()
	if err != nil {
		return Results{}, err
	}
	if len(clusterImages) == 0 {
		return Results{}, fmt.Errorf("no container images found in cluster")
	}

	images := make([]*Image, len(clusterImages))
	errored := make([]*ErroredImage, 0)
	g := new(errgroup.Group)
	for i, fullName := range clusterImages {
		i, fullName := i, fullName
		image, err := newImage(fullName)
		if err != nil {
			errored = append(errored, &ErroredImage{
				Image: fullName,
				Err:   err.Error(),
			})
			images[i] = nil
			continue
		}
		klog.V(5).Infof("Getting tags for %s", image.Name)
		g.Go(func() error {
			err := image.getTags()
			if err == nil {
				images[i] = image
			}
			return err
		})
	}
	if err := g.Wait(); err != nil {
		fmt.Println("Got an error when getting remote tags:", err)
		// return Results{}, err
	}
	for _, image := range images {
		if image == nil {
			continue
		}
		image.parseTags()
		err := image.populateNewest()
		if err != nil {
			return Results{}, err
		}
	}
	return Results{
		Images:    images,
		ErrImages: errored,
	}, nil
}

// getContainerImages fetches all pods and returns a slice of container images
func (c *Client) getContainerImages() ([]string, error) {
	klog.V(3).Infof("Getting container images from pods")

	k := c.Kube.Client
	pods, err := k.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "getting all pods")
	}

	imagesFound := make([]string, 0)
	for _, pod := range pods.Items {
		if len(pod.Spec.InitContainers) > 0 {
			for _, container := range pod.Spec.InitContainers {
				if container.Image != "" {
					imagesFound = append(imagesFound, container.Image)
				}
			}
		}
		for _, container := range pod.Spec.Containers {
			if container.Image != "" {
				imagesFound = append(imagesFound, container.Image)
			}
		}
	}
	imagesFound = removeDuplicateStr(imagesFound)
	return imagesFound, nil
}

func removeDuplicateStr(strSlice []string) []string {
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

func newImage(fullImageTag string) (*Image, error) {
	klog.V(3).Infof("Creating image object for %s", fullImageTag)

	var (
		err     error
		repo    = ""
		currTag = "latest"
		image   = new(Image)
	)

	splitString := strings.Split(fullImageTag, ":")
	if len(splitString) > 0 {
		repo = splitString[0]
		if len(splitString) > 1 {
			currTag = splitString[1]
		}
	}

	re, err := regexp.Compile(`^v[0-9]+.*$`)
	if err != nil {
		return nil, errors.Wrap(err, "failed to compile regex")
	}
	if re.MatchString(currTag) {
		image.Prefix = "v"
	}
	currTag = strings.TrimPrefix(currTag, image.Prefix)
	image.Name = repo
	ver, verString, strict := parseTagString(currTag)
	image.StrictSemver = strict
	image.Current = &Tag{
		version: ver,
		Value:   verString,
	}
	image.repo, err = name.NewRepository(repo)
	if err != nil {
		return nil, err
	}

	return image, nil
}

func (i *Image) getTags() error {
	tags, err := remote.List(i.repo, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return err
	}
	i.allTags = tags
	return nil
}

func (i *Image) parseTags() {
	for _, tag := range i.allTags {
		if i.Prefix != "" {
			tag = strings.TrimPrefix(tag, i.Prefix)
		}
		ver, verString, _ := parseTagString(tag)
		if ver != nil {
			i.semverTags = append(i.semverTags, ver)
		} else {
			i.nonSemverTags = append(i.nonSemverTags, verString)
		}
	}
}

func (i *Image) populateNewest() error {
	if i == nil || i.Current.version == nil {
		return nil
	}
	klog.V(3).Infof("Populating newest tags for %s", i.Name)
	newerTags := make([]*version.Version, 0)
	constraint, err := version.NewConstraint(fmt.Sprintf("> %s", i.Current.version.String()))
	if err != nil {
		return errors.Wrap(err, "failed to create constraint")
	}
	// The goal of the filter below is to find things like "1.2.3-alpine" or "1.2.3-buster" and make sure we only give upgrade suggestions that match.
	// There may be unintended consequenses from this. We do our best by ignoring actual common pre-release tags in the variable preReleaseIgnore.
	// This means if you are currently running 1.2.3-beta.0 the filter will not limit the upgrade suggestions to other beta releases.
	filter := i.Current.version.Prerelease()
	if err != nil {
		return err
	}
	for _, tag := range i.semverTags {
		if tag.Prerelease() != filter && preReleaseRegex(preReleaseIgnore, tag.Prerelease()) {
			continue
		}
		if constraint.Check(tag) {
			if tag.Major() > i.Current.version.Major()+10 {
				continue
			}
			newerTags = append(newerTags, tag)
		}
	}
	sort.Sort(sort.Reverse(version.Collection(newerTags)))
	if len(newerTags) > 0 {
		i.Newest = &Tag{
			version: newerTags[0],
			Value:   newerTags[0].String(),
		}
	}
	if i.Current.version.Major() > 0 {
		newerMinorConstraint, err := version.NewConstraint(fmt.Sprintf("%d.x.x", i.Current.version.Major()))
		if err != nil {
			return errors.Wrap(err, "failed to create minor version constraint")
		}
		for _, tag := range newerTags {
			if newerMinorConstraint.Check(tag) {
				i.NewestMinor = &Tag{
					version: tag,
					Value:   tag.String(),
				}
				break
			}
		}
	}
	newerPatchConstraint, err := version.NewConstraint(fmt.Sprintf("%d.%d.x", i.Current.version.Major(), i.Current.version.Minor()))
	if err != nil {
		return errors.Wrap(err, "failed to create patch version constraint")
	}
	for _, tag := range newerTags {
		if newerPatchConstraint.Check(tag) {
			i.NewestPatch = &Tag{
				version: tag,
				Value:   tag.String(),
			}
			break
		}
	}
	return nil
}

func parseTagString(versionString string) (*version.Version, string, bool) {
	strictV, err := version.StrictNewVersion(versionString)
	if err != nil {
		v, err := version.NewVersion(versionString)
		if err != nil {
			return nil, versionString, false
		}
		return v, versionString, false
	}
	return strictV, versionString, true
}

func preReleaseRegex(strings []string, prerelease string) bool {
	for _, str := range strings {
		reg, err := regexp.Compile(fmt.Sprintf(`^%s.*`, str))
		if err != nil {
			klog.Errorf("Failed to compile regex for %s", str)
			continue
		}
		if reg.MatchString(prerelease) {
			return true
		}
	}
	return false
}
