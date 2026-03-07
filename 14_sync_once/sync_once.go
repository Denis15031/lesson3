package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Connection struct {
	ID        int
	CreatedAt time.Time
}

type Database struct {
	once sync.Once
	conn *Connection
}

// возвращает подключение, инициализируя его при первом вызове
func (db *Database) GetConnection() *Connection {
	db.once.Do(func() {
		fmt.Println("Инициализация подключения к БД...")
		time.Sleep(500 * time.Millisecond)

		db.conn = &Connection{
			ID:        1,
			CreatedAt: time.Now(),
		}

		fmt.Printf("Подключение создано (ID: %d, Время: %v)\n",
			db.conn.ID, db.conn.CreatedAt.Format("15:04:05"))
	})
	return db.conn
}

// возвращает информацию о подключении (для демонстрации)
func (db *Database) GetConnectionInfo() string {
	if db.conn == nil {
		return "Подключение не инициализировано"
	}
	return fmt.Sprintf("Подключение #%d, создано в %v", db.conn.ID, db.conn.CreatedAt.Format("15:04:05"))
}

func main() {
	db := &Database{}
	var wg sync.WaitGroup
	numGoroutine := 10

	fmt.Println("Запуск горутин...\n")

	for i := 1; i <= numGoroutine; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

			conn := db.GetConnection()
			fmt.Printf("Горутина %d: получила подключение #%d\n", id, conn.ID)
		}(i)
	}
	wg.Wait()

	fmt.Printf("Итог: %s\n", db.GetConnectionInfo())
	fmt.Println("Все горутины завершили работу!")
}
