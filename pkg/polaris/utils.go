package polaris

import (
	"encoding/base64"
	b64 "encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
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
