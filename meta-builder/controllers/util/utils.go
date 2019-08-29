package util

import (
	"fmt"
	"io/ioutil"
	"net/http"
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
func GetBaseYaml(appName string, version string) (string, error) {
	// only fetch the location of the latest if the version in the spec is not given
	if 0 == len(version) {
		latestBaseYamlURL := fmt.Sprintf("https://raw.githubusercontent.com/blackducksoftware/releases/master/%s/latest", appName)
		latestArrayOfByte, err := HTTPGet(latestBaseYamlURL)
		if err != nil {
			return "", err
		}
		version = string(latestArrayOfByte)
	}

	versionBaseYamlURL := fmt.Sprintf("https://raw.githubusercontent.com/blackducksoftware/releases/master/%s/%s/%s_base.yaml", appName, version, appName)
	versionBaseYamlAsByteArray, err := HTTPGet(versionBaseYamlURL)
	if err != nil {
		return "", err
	}
	return string(versionBaseYamlAsByteArray), nil
}
