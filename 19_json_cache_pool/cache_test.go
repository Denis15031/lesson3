package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSetGet(t *testing.T) {
	c := NewObjectCache(2 * time.Second)
	defer c.Close()

	c.Set("key", "value")
	v, ok := c.Get("key")
	if !ok || v != "value" {
		t.Errorf("expected value, got %v", v)
	}
}

func TestTTLExpiration(t *testing.T) {
	c := NewObjectCache(100 * time.Millisecond)
	defer c.Close()

	c.Set("expiring", 42)
	time.Sleep(150 * time.Millisecond)
	_, ok := c.Get("expiring")
	if ok {
		t.Error("expected key to expire")
	}
}

func TestToJSON(t *testing.T) {
	c := NewObjectCache(5 * time.Second)
	defer c.Close()

	c.Set("a", 1)
	c.Set("b", "test")

	data, err := c.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty JSON")
	}
	// Простая проверка валидности JSON
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Errorf("invalid JSON: %v", err)
	}
}

func TestConcurrentAccess(t *testing.T) {
	c := NewObjectCache(10 * time.Second)
	defer c.Close()

	done := make(chan bool, 100) // буферизированный канал
	for i := 0; i < 100; i++ {
		go func(n int) {
			key := "key" + string(rune('0'+n%10))
			c.Set(key, n)
			c.Get(key)
			done <- true // ✅ Исправлено: было tr...ue
		}(i)
	}
	for i := 0; i < 100; i++ {
		<-done
	}
}

func TestDelete(t *testing.T) {
	c := NewObjectCache(5 * time.Second)
	defer c.Close()

	c.Set("del", 123)
	c.Delete("del")
	if _, ok := c.Get("del"); ok {
		t.Error("expected key to be deleted")
	}
}
