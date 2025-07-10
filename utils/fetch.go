package utils

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

var client = &http.Client{
	Timeout: 10 * time.Second,
}

func GetFile(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logrus.Errorf("Failed to fetch file from %s: %s", url, resp.Status)
		return nil, errors.New(resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("Failed to read response body from %s: %v", url, err)
		return nil, err
	}

	return data, nil
}
