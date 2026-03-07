package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

type ConfigManager struct {
	once   sync.Once
	config map[string]string
	err    error
}

func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		config: make(map[string]string),
	}
}

// загружает конфигурацию один раз (ленивая инициализация)
func (cm *ConfigManager) LoadConfig() error {
	cm.once.Do(func() {
		cm.config = map[string]string{
			"app_name":  "MyApp",
			"port":      "8080",
			"log_level": "debug",
			"db_host":   "localhost",
			"db_port":   "5432",
		}

		if envPort := os.Getenv("APP_PORT"); envPort != "" {
			cm.config["port"] = envPort
		}
		fmt.Println("Конфигурация загружена")
	})
	return cm.err
}

// возвращает значение конфигурации по ключу
func (cm *ConfigManager) Get(key string) string {
	if err := cm.LoadConfig(); err != nil {
		return ""
	}
	return cm.config[key]
}

// возвращает значение или дефолт, если ключ не найден
func (cm *ConfigManager) GetOrDefault(key, defaultValue string) string {
	if val := cm.Get(key); val != "" {
		return val
	}
	return defaultValue
}

// выводит все загруженные параметры
func (cm *ConfigManager) PrintConfig() {
	if err := cm.LoadConfig(); err != nil {
		fmt.Printf("Ошибка загрузки конфигурации: %v\n", err)
		return
	}

	fmt.Println("Текущая конфигурация:")
	fmt.Println(strings.Repeat("-", 30))
	for k, v := range cm.config {
		fmt.Printf("%-15s : %s\n", k, v)
	}
	fmt.Println(strings.Repeat("-", 30))

}

// проверяет, загружена ли конфигурация
func (cm *ConfigManager) IsLoaded() bool {
	_, loaded := cm.config["app_name"]
	return loaded
}

func main() {
	configManager := NewConfigManager()
	var wg sync.WaitGroup

	// Симулируем 10 горутин, запрашивающих конфиг одновременно
	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Каждая горутина запрашивает конфиг
			appName := configManager.Get("app_name")
			port := configManager.Get("port")

			fmt.Printf("Горутина %d: app=%s, port=%s\n", id, appName, port)
		}(i)
	}

	wg.Wait()

	// Выводим итоговую конфигурацию
	configManager.PrintConfig()

	fmt.Println("Конфигурация успешно загружена и использована!")
}
