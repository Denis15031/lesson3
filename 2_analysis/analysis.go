package main

import (
	"time"
)

func worker() chan int {
	ch := make(chan int)
	go func() {
		time.Sleep(3 * time.Second)
		ch <- 42
	}()
	return ch
}
func main() {
	timeStart := time.Now()

	//запускаем воркеры сразу
	ch1 := worker()
	ch2 := worker()

	//теперь читаем из обоих, теперь они работают параллельно
	_, _ = <-ch1, ch2

	// _, _ = <-worker(), <-worker() // поэтому и выведется 6, последовательное выполнение слева - направо. Воркеры запускаются последовательно, а не параллельно
	println(int(time.Since(timeStart).Seconds())) // 6 // выведется примерно за 7-8 секунд
}

// после исправления выведет 3 (примерно за ~3 секунды)

// Можно было бы через WaitGroup, но так проще
//Так же можно было бы собирать результаты в слайс, но это ещё более заморочисто нежели waitGroup
