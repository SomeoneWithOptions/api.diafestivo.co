package giphy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func FetchGifURL() *string {
	KEY := os.Getenv("GIPHY_KEY")
	var gif Gif
	GIPHY_QUERY := fmt.Sprintf("https://api.giphy.com/v1/gifs/random?api_key=%v&tag=celebrate&rating=g", KEY)

	res, err := http.Get(GIPHY_QUERY)

	if err != nil {
		panic("Error Making HTTP request to giphy: " + err.Error())
	}
	defer res.Body.Close()
	json.NewDecoder(res.Body).Decode(&gif)

	return &gif.Data.Images.Original.URL
}
