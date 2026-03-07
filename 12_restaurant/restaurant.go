package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Restaurant struct {
	mu             sync.Mutex
	cond           *sync.Cond
	totalTables    int
	occupiedTables int
	tables         []bool
	closed         bool
}

func NewRestaurant(totalTables int) *Restaurant {
	r := &Restaurant{
		totalTables: totalTables,
		tables:      make([]bool, totalTables),
	}
	r.cond = sync.NewCond(&r.mu)
	return r
}

// занимает столик, блокируется если все заняты
func (r *Restaurant) OccupyTable(guestID int) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	//Ждём пока не появится свободный столик или ресторан не закроется
	for r.occupiedTables >= r.totalTables && !r.closed {
		fmt.Printf("[Гость %d] Все столики заняты, ожидаем в очереди...\n", guestID)
		r.cond.Wait()
	}

	// Если ресторан закрыт во время ожидания
	if r.closed {
		fmt.Printf("[Гость %d] Ресторан закрыт, уходим", guestID)
		return -1
	}

	//Находим первый свободный столик
	for i := 0; i < r.totalTables; i++ {
		if !r.tables[i] {
			r.tables[i] = true
			r.occupiedTables++
			fmt.Printf("[Гость %d] Занял столик #%d (всего занято: %d/%d)\n", guestID, i+1, r.occupiedTables, r.totalTables)
			return i
		}
	}
	return -1
}

// освобождает столик и уведомляет ожидающих
func (r *Restaurant) ReleaseTable(guestID int, tableNum int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if tableNum >= 0 && tableNum < r.totalTables && r.tables[tableNum] {
		r.tables[tableNum] = false
		r.occupiedTables--
		fmt.Printf("[Гость %d] Освободил столик #%d (всего занято: %d/%d)\n", guestID, tableNum+1, r.occupiedTables, r.totalTables)

		r.cond.Signal()
	}
}

// закрывает ресторан
func (r *Restaurant) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.closed = true
	fmt.Println("[Ресторан] Закрыт, обслуживаем оставшихся гостей...")
	r.cond.Broadcast()
}

// проверяет закрыт ли ресторан
func (r *Restaurant) IsClosed() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.closed
}

func main() {
	restaurant := NewRestaurant(5)
	var wg sync.WaitGroup

	// Симуляция 15 гостей
	numGuests := 15
	wg.Add(numGuests)

	for i := 1; i <= numGuests; i++ {
		go func(guestID int) {
			defer wg.Done()

			// Случайная задержка перед приходом
			time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

			// Занимаем столик
			tableNum := restaurant.OccupyTable(guestID)
			if tableNum == -1 {
				return
			}

			// Симуляция посещения (1-3 секунды)
			diningTime := time.Duration(1+rand.Intn(2)) * time.Second
			time.Sleep(diningTime)

			// Освобождаем столик
			restaurant.ReleaseTable(guestID, tableNum)
		}(i)
	}

	// Закрываем ресторан через 8 секунд
	go func() {
		time.Sleep(8 * time.Second)
		restaurant.Close()
	}()

	wg.Wait()
	fmt.Println("[Ресторан] Все гости обслужены, работа завершена!")
}
