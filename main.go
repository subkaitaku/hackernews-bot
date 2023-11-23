package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func main() {
	http.HandleFunc("/", handler)

	port := 8080
	fmt.Printf("Server listening on :%d...\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
}

type ItemDetail struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	const topStoriesURL = "https://hacker-news.firebaseio.com/v0/topstories.json"
	stRes, err := http.Get(topStoriesURL)
	if err != nil {
		fmt.Println("error!!", err)
		return
	}
	defer stRes.Body.Close()

	body, err := io.ReadAll(stRes.Body)
	if err != nil {
		fmt.Println("body parse error!!: ", err)
		return
	}
	var storyNums []int
	err = json.Unmarshal([]byte(body), &storyNums)
	if err != nil {
		fmt.Println("error!!", err)
		return
	}

	var result []map[string]string
	for _, value := range storyNums[:10] {
		detailURl := "https://hacker-news.firebaseio.com/v0/item/" + strconv.Itoa(value) + ".json"
		detailRes, err := http.Get(detailURl)
		if err != nil {
			fmt.Println("error!!", err)
			return
		}
		defer detailRes.Body.Close()

		detailBody, err := io.ReadAll(detailRes.Body)
		if err != nil {
			fmt.Println("body parse error!!: ", err)
			return
		}

		var itemDetail ItemDetail
		err = json.Unmarshal(detailBody, &itemDetail)
		if err != nil {
			fmt.Println("json parse error!: ", err)
			return
		}
		result = append(result, map[string]string{"title": itemDetail.Title, "url": itemDetail.URL})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
