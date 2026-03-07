package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type RequestData struct {
	UserID    int64             `json:"user_id"`
	Action    string            `json:"action"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Tags      []string          `json:"tags,omitempty"`
}

// очищает все поля объекта перед возвратом в пул
func (r *RequestData) Reset() {
	r.UserID = 0
	r.Action = ""
	r.Timestamp = time.Time{}

	//// Очищаем map, но сохраняем сам объект
	if r.Metadata != nil {
		for k := range r.Metadata {
			delete(r.Metadata, k)
		}
	} else {
		r.Metadata = make(map[string]string)
	}

	//// Очищаем slice, сохраняя базовый массив (capacity)
	if r.Tags != nil {
		r.Tags = r.Tags[:0]
	} else {
		r.Tags = make([]string, 0, 8)
	}
}

// пул для переиспользования объектов RequestData
var requestDataPool = sync.Pool{
	New: func() interface{} {
		return &RequestData{
			Metadata: make(map[string]string, 8),
			Tags:     make([]string, 0, 8),
		}
	},
}

// получает объект из пула
func getRequest() *RequestData {
	return requestDataPool.Get().(*RequestData)
}

// возвращает объект в пул после очистки
func putRequest(r *RequestData) {
	r.Reset()
	requestDataPool.Put(r)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	data := getRequest()
	defer putRequest(data)

	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if data.Timestamp.IsZero() {
		data.Timestamp = time.Now()
	}

	responce := map[string]interface{}{
		"status":     "success",
		"user_id":    data.UserID,
		"action":     data.Action,
		"processed":  data.Timestamp,
		"tags_count": len(data.Tags),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responce)
}

// эндпоинт для демонстрации статистики
func handleStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"pool_active": "sync.Pool manages objects automatically",
		"memory":      "reduced GC pressure through object reuse",
		"thread_safe": "yes, sync.Pool is concurrent-safe",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func main() {
	http.HandleFunc("/", handleRequest)
	http.HandleFunc("/stats", handleStats)

	fmt.Println("Сервер запущен на порту :8080")
	fmt.Println("Тестовые запросы:")
	fmt.Println(" curl -X POST http://localhost:8080 -d '{\"user_id\":123,\"action\":\"login\",\"tags\":[\"web\",\"mobile\"]}'")
	fmt.Println(" curl http://localhost:8080/stats")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Ошибка сервера: %v\n", err)
	}
}
