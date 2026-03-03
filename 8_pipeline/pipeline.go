package main

import (
	"fmt"
	"sync"
	"time"
)

// добавляет префикс
func Parse(in <-chan string) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)
		for data := range in {
			out <- "parsed -" + data
		}
	}()

	return out
}

// распределяет данные по N каналам
func Split(in <-chan string, n int) []<-chan string {
	channels := make([]chan string, n)
	for i := 0; i < n; i++ {
		channels[i] = make(chan string)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer func() {
			for _, ch := range channels {
				close(ch)
			}
		}()

		idx := 0
		for data := range in {
			channels[idx] <- data
			idx = (idx + 1) % n
		}
	}()

	result := make([]<-chan string, n)
	for i, ch := range channels {
		result[i] = ch
	}

	return result
}

// обрабатывает данные из N каналов и объединяет
func Send(channels []<-chan string) <-chan string {
	out := make(chan string)
	var wg sync.WaitGroup

	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan string) {
			defer wg.Done()
			for data := range c {
				out <- "sent -" + data
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

//генератор входных данных
func generateData(data []string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for _, d := range data {
			out <- d
			time.Sleep(30 * time.Millisecond) // Имитация задержки
		}
	}()
	return out
}

func main() {
	// Входные данные
	rawData := []string{"alpha", "beta", "gamma", "delta", "epsilon"}
	input := generateData(rawData)

	//Строим конвейер
	parsed := Parse(input)
	splitChannels := Split(parsed, 3)
	results := Send(splitChannels)

	// Вывод результатов
	fmt.Println("Результаты конвейера:")
	fmt.Println("---------")
	for result := range results {
		fmt.Println("Успешно", result)
	}
	fmt.Println("---------")
	fmt.Println("Конвейер завершён")
}
