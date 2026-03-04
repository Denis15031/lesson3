package main

import (
	"fmt"
	"sync"
)

type SafeCache struct {
	mu   sync.RWMutex
	data map[string]string
}

// создает новый экземпляр кэша
func NewSafeCache() *SafeCache {
	return &SafeCache{
		data: make(map[string]string),
	}
}

func (c *SafeCache) Set(key string, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

func (c *SafeCache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists := c.data[key]
	return value, exists
}

func (c *SafeCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

func (c *SafeCache) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.data)
}

func (c *SafeCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]string)
}

func main() {
	fmt.Println("Тест кэша")

	cache := NewSafeCache()
	var wg sync.WaitGroup
	numWorkers := 10
	opsPerWorker := 100

	fmt.Println("Запуск записи из 10 горутин")
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerWorker; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				value := fmt.Sprintf("value_%d_%d", id, j)
				cache.Set(key, value)
			}
		}(i)
	}

	fmt.Println("Запуск чтения из 10 горутин")
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerWorker; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				cache.Get(key)
			}
		}(i)
	}

	wg.Wait()

	fmt.Printf("Завершено. Размер кэша:%d\n", cache.Size())

	fmt.Println("Демонстрация")

	cache.Set("name", "Vasya")
	cache.Set("age", "50")
	cache.Set("sity", "Novosibirsk")

	if val, ok := cache.Get("name"); ok {
		fmt.Printf("name = %s\n", val)
	}

	fmt.Printf("Текущий размер: %d\n", cache.Size())

	cache.Delete("age")
	if _, ok := cache.Get("age"); !ok {
		fmt.Println("Ключ 'age' удален")
	}

	cache.Clear()
	fmt.Printf("После очистки размер: %d\n", cache.Size())

	fmt.Println("Стресс-тест на гонки")
	runRaceTest(cache)
	fmt.Println("Стресс-тест завершён без гонок")
}

// интенсивный тест на параллельный доступ
func runRaceTest(cache *SafeCache) {
	var wg sync.WaitGroup
	workers := 50
	ops := 500

	for i := 0; i < workers; i++ {
		wg.Add(3)

		// Запись
		go func(id int) {
			defer wg.Done()
			for j := 0; j < ops; j++ {
				cache.Set(fmt.Sprintf("race_%d_%d", id, j), fmt.Sprintf("val_%d", j))
			}
		}(i)

		// Чтение
		go func(id int) {
			defer wg.Done()
			for j := 0; j < ops; j++ {
				cache.Get(fmt.Sprintf("race_%d_%d", id, j))
			}
		}(i)

		// Удаление
		go func(id int) {
			defer wg.Done()
			for j := 0; j < ops; j++ {
				cache.Delete(fmt.Sprintf("race_%d_%d", id, j))
			}
		}(i)
	}

	wg.Wait()
}
