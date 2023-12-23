package giphy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func GetGifURL() string {
	KEY := os.Getenv("GIPHY_KEY")
	GIPHY_QUERY := fmt.Sprintf("https://api.giphy.com/v1/gifs/random?api_key=%v&tag=celebrate&rating=g", KEY)
	res, err := http.Get(GIPHY_QUERY)
	if err != nil {
		panic("Error Making HTTP request")
	}
	defer res.Body.Close()
	resBytes, _ := io.ReadAll(res.Body)
	var gif Gif
	json.Unmarshal(resBytes, &gif)
	return gif.Data.Images.Original.URL
}
