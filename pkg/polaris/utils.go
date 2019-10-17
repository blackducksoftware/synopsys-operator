package polaris

import (
	"encoding/base64"
	b64 "encoding/base64"
	"fmt"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"regexp"
)

// HTTPGet returns the http response for the api
func HTTPGet(url string) (content []byte, err error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		proxyURL, _ := http.ProxyFromEnvironment(response.Request)
		if proxyURL != nil {
			return nil, fmt.Errorf("failed to fetch %s using proxy %s | %s", response.Request.URL.String(), proxyURL.String(), response.Status)
		}
		return nil, fmt.Errorf("failed to fetch %s | %s", response.Request.URL.String(), response.Status)
	}
	return ioutil.ReadAll(response.Body)
}

// GetBaseYaml returns the base yaml as string for the given app and version
func GetBaseYaml(baseurl string, appName string, version string, fileName string) (string, error) {
	// only fetch the location of the latest if the version in the spec is not given
	url, err := url.Parse(baseurl)
	if err != nil {
		return "", err
	}

	url.Path = path.Join(url.Path, appName, version, fileName)

	return downloadAndConvertYamlToByteArray(url.String())
}

func downloadAndConvertYamlToByteArray(url string) (string, error) {
	versionBaseYamlAsByteArray, err := HTTPGet(url)
	if err != nil {
		return "", err
	}
	return string(versionBaseYamlAsByteArray), nil
}

// EncodeStringToBase64 will return encoded string to base64
func EncodeStringToBase64(str string) string {
	return b64.StdEncoding.EncodeToString([]byte(str))
}

// Base64Encode will return an encoded string using a URL-compatible base64 format
func Base64Encode(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}

// Base64Decode will return a decoded string using a URL-compatible base64 format;
// decoding may return an error, which you can check if you don’t already know the input to be well-formed.
func Base64Decode(data string) (string, error) {
	uDec, err := base64.URLEncoding.DecodeString(data)
	return string(uDec), err
}

func updateRegistry(obj map[string]runtime.Object, registry string) (map[string]runtime.Object, error) {
	for _, v := range obj {
		if podspec := findPodSpec(reflect.ValueOf(v)); podspec != nil {
			if err := updateContainersImage(*podspec, registry); err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}

func findPodSpec(t reflect.Value) *corev1.PodSpec {
	podSpecType := reflect.TypeOf(corev1.PodSpec{})

	switch t.Kind() {
	case reflect.Ptr:
		return findPodSpec(t.Elem())
	case reflect.Struct:
		if t.Type() == podSpecType && t.CanInterface() {
			podSpec, _ := t.Interface().(corev1.PodSpec)
			return &podSpec
		}
		for i := 0; i < t.NumField(); i++ {
			if podSpec := findPodSpec(t.Field(i)); podSpec != nil {
				return podSpec
			}
		}
	case reflect.Array, reflect.Slice:
		for i := 0; i < t.Len(); i++ {
			if podSpec := findPodSpec(t.Index(i)); podSpec != nil {
				return podSpec
			}
		}
	case reflect.Map:
		for _, key := range t.MapKeys() {
			if podSpec := findPodSpec(t.MapIndex(key)); podSpec != nil {
				return podSpec
			}
		}
	}

	return nil
}

func updateContainersImage(podSpec corev1.PodSpec, registry string) error {
	for containerIndex, container := range podSpec.Containers {
		newImage, err := generateNewImage(container.Image, registry)
		if err != nil {
			return err
		}
		podSpec.Containers[containerIndex].Image = newImage
	}

	for initContainerIndex, initContainer := range podSpec.InitContainers {
		newImage, err := generateNewImage(initContainer.Image, registry)
		if err != nil {
			return err
		}
		podSpec.InitContainers[initContainerIndex].Image = newImage
	}
	return nil
}

func generateNewImage(currentImage string, registry string) (string, error) {
	imageTag, err := getImageAndTag(currentImage)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", registry, imageTag), nil
}

func getImageAndTag(image string) (string, error) {
	r := regexp.MustCompile(`^(|.*/)([a-zA-Z_0-9-.:]+)$`)
	groups := r.FindStringSubmatch(image)
	if len(groups) < 3 && len(groups[2]) == 0 {
		return "", fmt.Errorf("couldn't find image and tags in [%s]", image)
	}
	return groups[2], nil
}
