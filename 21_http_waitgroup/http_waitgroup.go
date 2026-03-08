package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

func fetchURL(url string) error {
	_, err := http.Get(url)
	return err
}

func main() {
	urls := []string{
		"https://www.lamoda.ru",
		"https://www.yandex.ru",
		"https://www.mail.ru",
		"https://www.google.ru",
	}

	var wg sync.WaitGroup

	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()

			u = strings.TrimSpace(u)

			fmt.Printf("Fetching %s...\n", u)
			err := fetchURL(u)
			if err != nil {
				fmt.Printf("Error fetching %s:%v\n", u, err)
				return
			}
			fmt.Printf("Fetched %s\n", u)
		}(url)
	}
	fmt.Println("All request launched!")
	wg.Wait()
	fmt.Println("Program finished")
}
