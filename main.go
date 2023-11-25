package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
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

type TranslationResponse struct {
	Translations []Translation `json:"translations"`
}

type Translation struct {
	DetectedSourceLanguage string `json:"detected_source_language"`
	Text                   string `json:"text"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	const topStoriesURL = "https://hacker-news.firebaseio.com/v0/topstories.json"
	const transApiURL = "https://api-free.deepl.com/v2/translate"
	const pushMessageURL = "https://api.line.me/v2/bot/message/push"
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
	if len(body) == 0 {
		fmt.Println("could not get the top stories")
		return
	}

	var storyNums []int
	err = json.Unmarshal([]byte(body), &storyNums)
	if err != nil {
		fmt.Println("error!!", err)
		return
	}

	var transResult []map[string]string
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
		if len(detailBody) == 0 {
			fmt.Println("could not get the detail story")
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
		if len(transBody) == 0 {
			break
		}

		var transSt TranslationResponse
		err = json.Unmarshal([]byte(transBody), &transSt)
		if err != nil {
			fmt.Println("Error json Unmarshal:", err)
			return
		}

		transResult = append(transResult, map[string]string{
			"originalTitle":   itemDetail.Title,
			"translatedTitle": transSt.Translations[0].Text,
			"url":             itemDetail.URL,
		})
	}
	if len(transResult) == 0 {
		fmt.Println("translate failed!")
		return
	}

	// build message data for LINE Message API
	var messages []string
	for i, value := range transResult {
		m := "【" + strconv.Itoa(i+1) + "位】 " + "\n" + value["translatedTitle"] + "(" + value["originalTitle"] + ")" + "\n" + value["url"]
		messages = append(messages, m)
		messages = append(messages, "")
	}
	result := strings.Join(messages, "\n")
	data := map[string]interface{}{
		"to": os.Getenv("LINE_USER_ID"),
		"messages": []map[string]string{
			{
				"type": "text",
				"text": result,
			},
		},
	}
	jData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("error!", err)
		return
	}

	req, err := http.NewRequest("POST", pushMessageURL, bytes.NewBuffer(jData))
	if err != nil {
		fmt.Println("post request creating error: ", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("LINE_CHANNEL_TOKEN"))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making HTTP request:", err)
		return
	}
	fmt.Println(res)
}
