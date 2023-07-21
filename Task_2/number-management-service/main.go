package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ResponseData struct {
	Numbers []int `json:"numbers"`
}

func main() {
	http.HandleFunc("/numbers", handleNumbers)
	http.ListenAndServe(":3000", nil)
}

func handleNumbers(w http.ResponseWriter, r *http.Request) {
	urls, ok := r.URL.Query()["url"]
	if !ok || len(urls) == 0 {
		http.Error(w, "Missing 'url' query parameter", http.StatusBadRequest)
		return
	}

	responseCh := make(chan ResponseData)
	errorCh := make(chan error)

	for _, url := range urls {
		go fetchNumbers(url, responseCh, errorCh)
	}

	mergedNumbers := make(map[int]bool)

	for i := 0; i < len(urls); i++ {
		select {
		case resp := <-responseCh:
			for _, num := range resp.Numbers {
				mergedNumbers[num] = true
			}
		case err := <-errorCh:
			fmt.Printf("Error: %s\n", err)
		case <-time.After(500 * time.Millisecond):
			fmt.Println("Timeout: Ignoring a URL")
		}
	}

	uniqueNumbers := []int{}
	for num := range mergedNumbers {
		uniqueNumbers = append(uniqueNumbers, num)
	}

	// I have used bubble sort here
	for i := 0; i < len(uniqueNumbers)-1; i++ {
		for j := 0; j < len(uniqueNumbers)-i-1; j++ {
			if uniqueNumbers[j] > uniqueNumbers[j+1] {
				uniqueNumbers[j], uniqueNumbers[j+1] = uniqueNumbers[j+1], uniqueNumbers[j]
			}
		}
	}

	response := ResponseData{Numbers: uniqueNumbers}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func fetchNumbers(url string, responseCh chan<- ResponseData, errorCh chan<- error) {
	resp, err := http.Get(url)
	if err != nil {
		errorCh <- err
		return
	}
	defer resp.Body.Close()

	var responseData ResponseData
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		errorCh <- err
		return
	}

	responseCh <- responseData
}
