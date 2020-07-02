package installation

import (
	"fmt"
	"net/http"
)

func getHttpClient() *http.Client {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}
	return client
}

func doGet(url string) (int, error) {
	httpClient := getHttpClient()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("cannot create a new HTTP request: %v", err)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("cannot send HTTP request to %s: %v", url, err)
	}
	return resp.StatusCode, nil
}
