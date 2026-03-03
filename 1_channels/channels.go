package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {

	//Используем локальный источник случайных чисел, это избавит от гонок
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	//Встроенная мапа не потокобезопасная, добавляем для защиты доступа
	var mu sync.Mutex
	alreadyStored := make(map[int]struct{})
	capacity := 1000
	doubles := make([]int, 0, capacity)
	for i := 0; i < capacity; i++ {
		doubles = append(doubles, r.Intn(10)) // используем локальный инстанс
	}
	uniqueIDs := make(chan int, capacity)
	wg := sync.WaitGroup{}

	for i := 0; i < capacity; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()

			//Блокируем мьютекс перед доступом к общей памяти
			mu.Lock()
			if _, ok := alreadyStored[doubles[i]]; !ok {
				alreadyStored[doubles[i]] = struct{}{}
				mu.Unlock()

				uniqueIDs <- doubles[i]
			} else {
				mu.Unlock()
			}
		}()
	}

	// Чтобы не было дедлока, запускаем отдельную горутину, которая ждет завершения всех воркеров
	go func() {
		wg.Wait()
		close(uniqueIDs)
	}()

	for val := range uniqueIDs {
		fmt.Println(val)
	}
}
