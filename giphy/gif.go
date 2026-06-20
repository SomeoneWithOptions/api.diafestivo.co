package giphy

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

const maxResponseBytes = 1 << 20

var (
	ErrMissingAPIKey = errors.New("missing GIPHY_KEY")
	gifHTTPClient    = &http.Client{Timeout: 3 * time.Second}
)

func FetchGifURL() *string {
	gifURL, err := FetchGifURLContext(context.Background())
	if err != nil {
		return nil
	}
	return gifURL
}

func FetchGifURLContext(ctx context.Context) (*string, error) {
	key := os.Getenv("GIPHY_KEY")
	if key == "" {
		return nil, ErrMissingAPIKey
	}

	requestURL := url.URL{
		Scheme: "https",
		Host:   "api.giphy.com",
		Path:   "/v1/gifs/random",
	}
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

	var gif Gif
	if err := json.NewDecoder(io.LimitReader(res.Body, maxResponseBytes)).Decode(&gif); err != nil {
		return nil, err
	}

	gifURL := gif.Data.Images.Original.URL
	if gifURL == "" {
		return nil, errors.New("giphy response missing original image url")
	}

	return new(gifURL), nil
}
