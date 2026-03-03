package main

import (
	"fmt"
	"sync"
)

func main() {
	naturals := make(chan int)
	squares := make(chan int)

	var wg sync.WaitGroup

	//генератор случайных чисел
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(naturals)
		for i := 1; i <= 10; i++ {
			naturals <- i
		}
	}()

	//обработчик
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(squares)
		for n := range naturals {
			squares <- n * n
		}
	}()

	//вывод результата
	for val := range squares {
		fmt.Println(val)
	}
	wg.Wait()
}
