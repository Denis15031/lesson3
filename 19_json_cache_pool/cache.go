package main

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"
	"time"
)

// хранит значение и время истечения TTL
type item struct {
	value      interface{}
	expiration int64
}

// потокобезопасный кэш с TTL и оптимизированной сериализацией
type ObjectCache struct {
	mu     sync.RWMutex
	data   map[string]item
	ttl    time.Duration
	pool   sync.Pool // пул *bytes.Buffer для сериализации
	stopCh chan struct{}
}

// создаёт новый кэш с заданным TTL и запускает фоновую очистку
func NewObjectCache(ttl time.Duration) *ObjectCache {
	c := &ObjectCache{
		data:   make(map[string]item),
		ttl:    ttl,
		stopCh: make(chan struct{}),
		pool: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
	go c.cleanupLoop()
	return c
}

// добавляет объект в кэш
func (c *ObjectCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = item{
		value:      value,
		expiration: time.Now().Add(c.ttl).UnixNano(),
	}
}

// возвращает объект по ключу, если он не истёк
func (c *ObjectCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	it, ok := c.data[key]
	c.mu.RUnlock()

	if !ok || time.Now().UnixNano() > it.expiration {
		if ok {
			c.Delete(key) // ленивое удаление
		}
		return nil, false
	}
	return it.value, true
}

func (c *ObjectCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

// сериализует актуальные данные кэша в JSON с использованием пула буферов
func (c *ObjectCache) ToJSON() ([]byte, error) {
	c.mu.RLock()
	// Копируем только актуальные данные
	safeData := make(map[string]interface{}, len(c.data))
	now := time.Now().UnixNano()
	for k, v := range c.data {
		if now <= v.expiration {
			safeData[k] = v.value
		}
	}
	c.mu.RUnlock()

	// Берём буфер из пула
	buf := c.pool.Get().(*bytes.Buffer)
	buf.Reset()
	defer c.pool.Put(buf)

	// Сериализуем
	enc := json.NewEncoder(buf)
	if err := enc.Encode(safeData); err != nil {
		return nil, err
	}

	// Возвращаем копию данных, так как буфер будет возвращён в пул
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result, nil
}

// периодически удаляет устаревшие записи
func (c *ObjectCache) cleanupLoop() {
	ticker := time.NewTicker(c.ttl / 2)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopCh:
			return
		}
	}
}

// удаляет истёкшие записи
func (c *ObjectCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now().UnixNano()
	for k, v := range c.data {
		if now > v.expiration {
			delete(c.data, k)
		}
	}
}

// останавливает фоновую горутину очистки
func (c *ObjectCache) Close() {
	close(c.stopCh)
}

// Убедимся, что bytes.Buffer реализует io.Writer (для ясности)
var _ io.Writer = (*bytes.Buffer)(nil)
