package main

import (
	"fmt"
	"sync"
	"time"
)

type BoundedQueue struct {
	mu       sync.Mutex
	cond     *sync.Cond
	tasks    []interface{}
	capacity int
	closed   bool
}

func NewBoundedQueue(capacity int) *BoundedQueue {
	q := &BoundedQueue{
		tasks:    make([]interface{}, 0, capacity),
		capacity: capacity,
	}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *BoundedQueue) Put(task interface{}) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	for len(q.tasks) >= q.capacity && !q.closed {
		q.cond.Wait()
	}

	if q.closed {
		return false
	}

	q.tasks = append(q.tasks, task)
	q.cond.Signal()
	return true
}

func (q *BoundedQueue) Get() (interface{}, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for len(q.tasks) == 0 && !q.closed {
		q.cond.Wait()
	}

	if len(q.tasks) == 0 {
		return nil, false
	}

	task := q.tasks[0]
	q.tasks = q.tasks[1:]
	q.cond.Signal()
	return task, true
}

func (q *BoundedQueue) Shutdown() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.closed = true
	q.cond.Broadcast()
}

func (q *BoundedQueue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.tasks)
}

func (q *BoundedQueue) IsClosed() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.closed
}

func main() {
	fmt.Println("Очередь с ограниченной емкостью\n")

	queue := NewBoundedQueue(5)

	var producerWg sync.WaitGroup
	var consumerWg sync.WaitGroup
	var processed int
	var processedMu sync.Mutex

	// Продюсеры
	fmt.Println("Запуск 3 продюсеров...")
	for i := 0; i < 3; i++ {
		producerWg.Add(1)
		go func(id int) {
			defer producerWg.Done()
			for j := 0; j < 5; j++ {
				task := fmt.Sprintf("task-%d-%d", id, j)
				if ok := queue.Put(task); ok {
					fmt.Printf("[Продюсер %d] Добавлено: %s\n", id, task)
				}
				time.Sleep(20 * time.Millisecond)
			}
		}(i)
	}

	// Консьюмеры
	fmt.Println("Запуск 2 консьюмеров...")
	for i := 0; i < 2; i++ {
		consumerWg.Add(1)
		go func(id int) {
			defer consumerWg.Done()
			for {
				task, ok := queue.Get()
				if !ok {
					fmt.Printf("[Консьюмер %d] Очередь закрыта, завершение\n", id)
					return
				}
				fmt.Printf("[Консьюмер %d] Обработано: %v\n", id, task)
				processedMu.Lock()
				processed++
				processedMu.Unlock()
				time.Sleep(30 * time.Millisecond)
			}
		}(i)
	}

	// Ждём завершения продюсеров
	producerWg.Wait()
	fmt.Println("\nПродюсеры завершили работу")

	//Закрываем очередь - консьюмеры получат сигнал
	fmt.Println("Закрытие очереди...")
	queue.Shutdown()

	//Ждём завершения консьюмеров
	consumerWg.Wait()

	fmt.Println("Все горутины завершены корректно")
	fmt.Printf("Итог: обработано=%d, пуста=%v, закрыта=%v\n",
		processed, queue.Size() == 0, queue.IsClosed())
}
