package integrations

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
)

// NewGISClient builds the HTTP client used to reach the GIS backend.
func NewGISClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
	}
	return &http.Client{Transport: tr}
}

// FetchGeometry retrieves the geometry payload for a contract.
func FetchGeometry(baseURL, contractID string) ([]byte, error) {
	resp, err := NewGISClient().Get(baseURL + "/geometry/" + contractID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
