package main

import (
	"fmt"
	"log"
	"sync"
)

// Интерфейс для всех плагинов
type Plugin interface {
	Execute() string
}

// Управляет инициализацией и доступом к плагинам
type PluginManager struct {
	plugins map[string]*pluginEntry
	mu      sync.RWMutex
}

type pluginEntry struct {
	once   sync.Once
	initFn func() (Plugin, error)
	plugin Plugin
	err    error
}

// NewPluginManager создает новый менеджер плагинов
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]*pluginEntry),
	}
}

// RegisterPlugin регистрирует новый плагин
func (pm *PluginManager) RegisterPlugin(name string, initFn func() (Plugin, error)) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.plugins[name] = &pluginEntry{
		initFn: initFn,
	}
	log.Printf("Плагин %q зарегистрирован", name)
}

// GetPlugin возвращает инициализированный плагин
func (pm *PluginManager) GetPlugin(name string) (Plugin, error) {
	pm.mu.RLock()
	entry, exists := pm.plugins[name]
	pm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin %q not registered", name)

	}

	entry.once.Do(func() {
		log.Printf("Инициализация плагина %q...", name)

		entry.plugin, entry.err = entry.initFn()

		if entry.err != nil {
			log.Printf("Ошибка инициализации %q: %v", name, entry.err)
		} else {
			log.Printf("Плагин %q инициализирован", name)
		}
	})
	return entry.plugin, entry.err

}

// DemoPlugin реализация плагина
type DemoPlugin struct{}

func (p *DemoPlugin) Execute() string {
	return "DemoPlugin executed successfully!"
}

func initDemo() (Plugin, error) {
	// Имитация длительной инициализации
	// time.Sleep(500 * time.Millisecond)
	return &DemoPlugin{}, nil
}

// плагин, который всегда падает при инициализации
type BrokenPlugin struct{}

func (p *BrokenPlugin) Execute() string {
	return "Never called"
}

func initBroken() (Plugin, error) {
	return nil, fmt.Errorf("simulated initialization error")
}

func main() {
	pm := NewPluginManager()

	pm.RegisterPlugin("demo", initDemo)
	pm.RegisterPlugin("broken", initBroken)

	var wg sync.WaitGroup

	// Тестирование рабочего плагина
	fmt.Println("Тест: рабочий плагин 'demo' ")
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			p, err := pm.GetPlugin("demo")
			if err != nil {
				log.Printf("Goroutine %d error: %v", id, err)
				return
			}
			log.Printf("Goroutine %d: %s", id, p.Execute())
		}(i)
	}

	// Тестирование плагина с ошибкой
	fmt.Println("Тест: плагин с ошибкой 'broken' ")
	for i := 5; i < 7; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, err := pm.GetPlugin("broken")
			if err != nil {
				log.Printf("Goroutine %d error: %v", id, err)
			}
		}(i)
	}

	//запрос незарегистрированного плагина
	fmt.Println("Тест: незарегистрированный плагин")
	_, err := pm.GetPlugin("nonexistent")
	if err != nil {
		log.Printf("Ожидаемая ошибка: %v", err)
	}

	wg.Wait()
	fmt.Println("Все тесты завершены!")
}
