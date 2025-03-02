package requests

import (
	"net/http"
	"time"
)

func GetRequest(url string) (*http.Response, error) {
	client := http.Client{Timeout: time.Second * 5}

	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}

	return client.Do(req)
}
