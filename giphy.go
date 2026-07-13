package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const maxGiphyResponseBytes = 1 << 20

var (
	errMissingGiphyKey = errors.New("missing GIPHY_KEY")
	gifHTTPClient      = &http.Client{Timeout: 3 * time.Second}
)

type gifResponse struct {
	Data struct {
		Images struct {
			Original struct {
				URL string `json:"url"`
			} `json:"original"`
		} `json:"images"`
	} `json:"data"`
}

func fetchGifURLContext(ctx context.Context) (*string, error) {
	key := os.Getenv("GIPHY_KEY")
	if key == "" {
		return nil, errMissingGiphyKey
	}

	requestURL := url.URL{Scheme: "https", Host: "api.giphy.com", Path: "/v1/gifs/random"}
	query := requestURL.Query()
	query.Set("api_key", key)
	query.Set("tag", "celebrate")
	query.Set("rating", "g")
	requestURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return nil, err
	}

	res, err := gifHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("giphy returned status %d", res.StatusCode)
	}

	var gif gifResponse
	if err := json.NewDecoder(io.LimitReader(res.Body, maxGiphyResponseBytes)).Decode(&gif); err != nil {
		return nil, err
	}

	gifURL := gif.Data.Images.Original.URL
	if gifURL == "" {
		return nil, errors.New("giphy response missing original image url")
	}
	return new(gifURL), nil
}
