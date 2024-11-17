package giphy

import (
	"fmt"
	"net/http"
	"os"

	j "github.com/json-iterator/go"
)

func GetGifURL() *string {
	KEY := os.Getenv("GIPHY_KEY")
	var gif Gif
	GIPHY_QUERY := fmt.Sprintf("https://api.giphy.com/v1/gifs/random?api_key=%v&tag=celebrate&rating=g", KEY)
	res, err := http.Get(GIPHY_QUERY)
	if err != nil {
		panic("Error Making HTTP request")
	}
	defer res.Body.Close()
	j.NewDecoder(res.Body).Decode(&gif)

	gif_url := gif.Data.Images.Original.URL
	return &gif_url
}
