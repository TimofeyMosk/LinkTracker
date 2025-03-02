package requests

import (
	"io"
	"net/http"
	"time"
)

func DeleteRequest(url string, data io.Reader) (*http.Response, error) {
	client := http.Client{Timeout: time.Second * 5}

	req, err := http.NewRequest(http.MethodDelete, url, data)
	if err != nil {
		return nil, err
	}

	return client.Do(req)
}
