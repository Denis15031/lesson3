package main

import (
	"fmt"
	"sync"
	"time"
)

func mergeChannels(channels ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup

	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan int) {
			defer wg.Done()
			for val := range c {
				out <- val
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func main() {
	a := make(chan int)
	b := make(chan int)
	c := make(chan int)

	go func() {
		defer close(a)
		for i := 1; i <= 3; i++ {
			a <- i
			time.Sleep(100 * time.Millisecond)
		}
	}()

	go func() {
		defer close(b)
		for i := 4; i <= 6; i++ {
			b <- i
			time.Sleep(150 * time.Millisecond)
		}
	}()

	go func() {
		defer close(c)
		for i := 7; i <= 9; i++ {
			c <- i
			time.Sleep(120 * time.Millisecond)
		}
	}()

	merged := mergeChannels(a, b, c)
	for val := range merged {
		fmt.Println(val)
	}
}
