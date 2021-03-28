package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	baseURL = "http://www.omdbapi.com/"
	apiKey  = "c082261f"
)

type OMDBManager struct {
}

func (tmdb *OMDBManager) DownloadImageForMovie(name string) (string, error) {
	processedName := processTorrentName(name)
	endpoint := fmt.Sprintf("?apikey=%s&t=%s", apiKey, url.QueryEscape(processedName))
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", baseURL, endpoint), nil)
	if err != nil {
		return "", err
	}

	resp, err := executeHTTPRequest(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var movieSearchResponse MovieSearchResponse
	err = json.Unmarshal(responseData, &movieSearchResponse)
	if err != nil {
		return "", err
	}

	return movieSearchResponse.Poster, nil
}

func executeHTTPRequest(req *http.Request) (*http.Response, error) {
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func processTorrentName(name string) string {
	re := regexp.MustCompile(`(.+)(\.\d{4}|\.S\d{2})`)

	matches := re.FindStringSubmatch(name)
	if len(matches) < 2 {
		return name
	}

	processedName := matches[1]
	processedName = strings.ReplaceAll(processedName, ".", " ")

	return processedName
}
