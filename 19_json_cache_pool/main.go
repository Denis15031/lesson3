package main

import (
	"fmt"
	"time"
)

func main() {
	cache := NewObjectCache(5 * time.Second)
	defer cache.Close()

	// Добавляем данные
	cache.Set("user:1", map[string]string{"name": "Alice", "role": "admin"})
	cache.Set("user:2", map[string]string{"name": "Bob", "role": "user"})

	// Получаем объект
	if user, found := cache.Get("user:1"); found {
		fmt.Println("Найден:", user)
	}

	// Сериализуем в JSON
	jsonData, err := cache.ToJSON()
	if err != nil {
		fmt.Println("Ошибка сериализации:", err)
		return
	}
	fmt.Println("Кэш в JSON:", string(jsonData))

	// Ждём истечения TTL
	fmt.Println("Ожидаем 6 секунд для истечения TTL...")
	time.Sleep(6 * time.Second)

	// Проверяем снова
	_, found := cache.Get("user:1")
	fmt.Println("После TTL, user:1 найден?", found)
}
