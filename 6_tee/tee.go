package main

import (
	"fmt"
	"sync"
	"time"
)

// имитирует рекплику БД:читает из канала и записывает данные
func dbReplica(name string, in <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for data := range in {
		fmt.Printf("Запись в %s: %d\n", name, data)
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Printf("Реплика %s закрыта\n", name)
}

func tee(in <-chan int, outputs ...chan<- int) <-chan struct{} {
	done := make(chan struct{})
	var wg sync.WaitGroup

	for _, out := range outputs {
		wg.Add(1)
		go func(ch chan<- int) {
			defer wg.Done()
			for data := range in {
				ch <- data
			}
			close(ch)
		}(out)
	}

	go func() {
		wg.Wait()
		close(done)
	}()
	return done
}

func main() {
	input := make(chan int)

	replicas := []chan int{
		make(chan int),
		make(chan int),
		make(chan int),
	}

	var replicaWg sync.WaitGroup
	replicaWg.Add(len(replicas))

	for i, ch := range replicas {
		go dbReplica(fmt.Sprintf("replica-%d", i+1), ch, &replicaWg)
	}

	done := tee(input, replicas[0], replicas[1], replicas[2])

	go func() {
		for i := 1; i <= 5; i++ {
			fmt.Printf("Отправка: %d\n", i)
			input <- i
			time.Sleep(50 * time.Millisecond)
		}
		close(input)
	}()

	<-done
	replicaWg.Wait()

	fmt.Println("Все реплики синхронизированы")
}
