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

	res, _ := http.Get(GIPHY_QUERY)

	resBytes, _ := io.ReadAll(res.Body)

	var gif Gif
	json.Unmarshal(resBytes, &gif)

	return fmt.Sprintf("%v", gif.Data.Images.Original.URL)
}
