package util

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"

	b64 "encoding/base64"

	"k8s.io/apimachinery/pkg/runtime"
)

// HTTPGet returns the http response for the api
func HTTPGet(url string) (content []byte, err error) {
	response, err := http.Get(url)
	if err != nil {
		return
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("INVALID RESPONSE; status: %s", response.Status)
	}
	return ioutil.ReadAll(response.Body)
}

// GetBaseYaml returns the base yaml as string for the given app and version
func GetBaseYaml(appName string, version string, fileName string) (string, error) {
	// only fetch the location of the latest if the version in the spec is not given
	if 0 == len(version) {
		latestBaseYamlURL := fmt.Sprintf("https://raw.githubusercontent.com/blackducksoftware/releases/master/%s/latest", appName)
		latestArrayOfByte, err := HTTPGet(latestBaseYamlURL)
		if err != nil {
			return "", err
		}
		version = string(latestArrayOfByte)
	}

	if 0 == len(fileName) {
		return downloadAndConvertYamlToByteArray(fmt.Sprintf("https://raw.githubusercontent.com/blackducksoftware/releases/master/%s/%s/%s_base.yaml", appName, version, appName))
	}
	return downloadAndConvertYamlToByteArray(fmt.Sprintf("https://raw.githubusercontent.com/blackducksoftware/releases/master/%s/%s/%s_base.yaml", appName, version, fileName))
}

func downloadAndConvertYamlToByteArray(url string) (string, error) {
	versionBaseYamlAsByteArray, err := HTTPGet(url)
	if err != nil {
		return "", err
	}
	return string(versionBaseYamlAsByteArray), nil
}

func GetAuthObjectsList() []string {
	return []string{
		"ServiceAccount.auth-server",
		"Role.leaderElector",
		"RoleBinding.leaderElector",
		"Deployment.auth-server",
		"HorizontalPodAutoscaler.auth-server",
		"Service.auth-server",
	}
}

func GetAuthServerRuntimeObjects(objects map[string]runtime.Object) map[string]runtime.Object {
	authServerRuntimeObjects := make(map[string]runtime.Object, 0)
	for _, entry := range GetAuthObjectsList() {
		authServerRuntimeObjects[entry] = objects[entry]
	}
	return authServerRuntimeObjects
}

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

// GetResourceName returns the name of the resource
func GetResourceName(name string, appName string, defaultName string) string {
	if len(appName) == 0 {
		return fmt.Sprintf("%s-%s", name, defaultName)
	}

	if len(defaultName) == 0 {
		return fmt.Sprintf("%s-%s", name, appName)
	}

	return fmt.Sprintf("%s-%s-%s", name, appName, defaultName)
}
