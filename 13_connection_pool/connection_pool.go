package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Connection struct {
	ID int
}

type ConnectionPool struct {
	mu          sync.Mutex
	cond        *sync.Cond
	connections []*Connection
	available   []*Connection
	inUse       int
	maxActive   int
	closed      bool
}

// создаёт пул с заданным размером
func NewConnectionPool(maxActive int) *ConnectionPool {
	pool := &ConnectionPool{
		maxActive:   maxActive,
		connections: make([]*Connection, maxActive),
		available:   make([]*Connection, 0, maxActive),
	}
	pool.cond = sync.NewCond(&pool.mu)

	for i := 0; i < maxActive; i++ {
		conn := &Connection{ID: i + 1}
		pool.connections[i] = conn
		pool.available = append(pool.available, conn)
	}
	return pool
}

// получает свободное подключение (блокируется, если нет доступных)
func (p *ConnectionPool) Get() *Connection {
	p.mu.Lock()
	defer p.mu.Unlock()

	for len(p.available) == 0 && !p.closed {
		fmt.Printf("Все подключения заняты, ожидаем...\n")
		p.cond.Wait()
	}

	if p.closed || len(p.available) == 0 {
		return nil
	}

	conn := p.available[0]
	p.available = p.available[1:]
	p.inUse++

	fmt.Printf("Подключение #%d выдано (активно: %d/%d)\n", conn.ID, p.inUse, p.maxActive)

	return conn
}

// освобождает подключение и уведомляет ожидающих
func (p *ConnectionPool) Release(conn *Connection) {
	if conn == nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.available = append(p.available, conn)
	p.inUse--

	fmt.Printf("Подключение #%d возвращено (активно: %d/%d)\n", conn.ID, p.inUse, p.maxActive)

	p.cond.Signal()
}

// закрывает пул и будит все ожидающие горутины
func (p *ConnectionPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.closed = true
	fmt.Println("Пул подключений закрыт")
	p.cond.Broadcast()

}

func main() {
	pool := NewConnectionPool(3)
	var wg sync.WaitGroup

	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

			conn := pool.Get()
			if conn == nil {
				fmt.Printf("Горутина %d: не удалось получить подключение\n", id)
				return
			}

			queryTime := time.Duration(1+rand.Intn(2)) * time.Second
			fmt.Printf("Горутина %d: выполняет запрос (подключение #%d)...\n", id, conn.ID)
			time.Sleep(queryTime)

			pool.Release(conn)
			fmt.Printf("Горутина %d: запрос завершён\n", id)
		}(i)
	}

	wg.Wait()
	pool.Close()
	fmt.Println("Все запросы обработаны, работа завершена!")
}
