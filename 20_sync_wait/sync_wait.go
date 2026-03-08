package main

import (
	"fmt"
	"sync"
)

func main() {
	cnt := 100
	var wg sync.WaitGroup

	for i := 0; i < cnt; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			fmt.Println(n)
		}(i)
	}

	wg.Wait()
	fmt.Println("Все горутины завершены!")
}
