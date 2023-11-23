package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	http.HandleFunc("/", handler)

	port := 8080
	fmt.Printf("Server listening on :%d...\n", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
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
	const transApiURL = "https://api-free.deepl.com/v2/translate"
	const targetLang = "JA"

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

		// translate DEEPL en -> ja
		data := url.Values{}
		data.Set("text", itemDetail.Title)
		data.Set("target_lang", targetLang)
		req, err := http.NewRequest("POST", transApiURL, bytes.NewBufferString(data.Encode()))
		if err != nil {
			fmt.Println("post request creating error: ", err)
			return
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Authorization", "DeepL-Auth-Key "+os.Getenv("DEEPL_API_KEY"))

		client := &http.Client{}
		transRes, err := client.Do(req)
		if err != nil {
			fmt.Println("Error making HTTP request:", err)
			return
		}
		defer transRes.Body.Close()
		transBody, err := io.ReadAll(transRes.Body)
		if err != nil {
			fmt.Println("Error read trans request:", err)
			return
		}

		result = append(result, map[string]string{
			"originalTitle":   itemDetail.Title,
			"translatedTitle": string(transBody),
			"url":             itemDetail.URL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
