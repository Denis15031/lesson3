package main

import (
	"fmt"
	"sync"
	"time"
)

type Comment struct {
	ID      int
	UserID  int
	Content string
}
type User struct {
	ID    int
	Name  string
	Email string
}
type Session struct {
	ID        string
	UserID    int
	ExpiresAt time.Time
}
type Attachment struct {
	ID   int
	Name string
	URL  string
}

// система параллельной загрузки данных
type DataLoader struct {
	mu                  sync.Mutex
	comments            []Comment
	users               map[int]*User
	session             *Session
	attachments         []Attachment
	sessionID           string
	loadAttachmentsOnce sync.Once
}

// создаёт новый загрузчик
func NewDataLoader(sessionID string) *DataLoader {
	return &DataLoader{
		users:     make(map[int]*User),
		sessionID: sessionID,
	}
}

// асинхронная загрузка комментариев
func (dl *DataLoader) loadComments(wg *sync.WaitGroup, commentsReady chan bool) {
	defer wg.Done()
	fmt.Println(" [Горутина 1] Загрузка комментариев из БД...")
	time.Sleep(100 * time.Millisecond)

	dl.mu.Lock()
	dl.comments = []Comment{
		{ID: 1, UserID: 101, Content: "Отличная статья"},
		{ID: 2, UserID: 102, Content: "Спасибо за информацию"},
		{ID: 3, UserID: 103, Content: "Полезно"},
	}
	fmt.Printf(" [Горутина 1] Загружено %d комментариев\n", len(dl.comments))
	dl.mu.Unlock()

	close(commentsReady) // сигнал
}

// параллельная загрузка сессии (
func (dl *DataLoader) loadSession(wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println(" [Горутина 2] Загрузка данных сессии...")
	time.Sleep(80 * time.Millisecond)

	dl.mu.Lock()
	dl.session = &Session{
		ID:        dl.sessionID,
		UserID:    101,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	fmt.Printf("[Горутина 2] Сессия загружена: %s\n", dl.session.ID)
	dl.mu.Unlock()
}

// загрузка пользователей
func (dl *DataLoader) loadUsers(wg *sync.WaitGroup, commentsReady chan bool) {
	defer wg.Done()

	<-commentsReady // ждём сигнал о готовности комментариев
	fmt.Println("[Горутина 3] Комментарии получены, загрузка пользователей...")
	time.Sleep(100 * time.Millisecond)

	dl.mu.Lock()
	userIDs := make(map[int]bool)
	for _, c := range dl.comments {
		userIDs[c.UserID] = true
	}
	for uid := range userIDs {
		dl.users[uid] = &User{
			ID:    uid,
			Name:  fmt.Sprintf("User%d", uid),
			Email: fmt.Sprintf("user%d@example.com", uid),
		}
	}
	fmt.Printf("[Горутина 3]Загружено %d пользователей\n", len(dl.users))
	dl.mu.Unlock()
}

// условная загрузка вложений
func (dl *DataLoader) loadAttachments(wg *sync.WaitGroup) {
	defer wg.Done()

	dl.loadAttachmentsOnce.Do(func() {
		if dl.sessionID == "" {
			fmt.Println(" [Горутина 4] Session-ID пуст, пропускаем вложения")
			return
		}
		fmt.Println(" [Горутина 4] Загрузка вложений...")
		time.Sleep(70 * time.Millisecond)

		dl.mu.Lock()
		dl.attachments = []Attachment{
			{ID: 1, Name: "doc.pdf", URL: "https://cdn.example.com/doc.pdf"},
			{ID: 2, Name: "img.png", URL: "https://cdn.example.com/img.png"},
		}
		fmt.Printf("[Горутина 4]Загружено %d вложений\n", len(dl.attachments))
		dl.mu.Unlock()
	})
}

// координирует все загрузки
func (dl *DataLoader) LoadAll() {
	var wg sync.WaitGroup
	commentsReady := make(chan bool)

	fmt.Println("Запуск параллельной загрузки данных...\n")

	// Параллельно: комментарии + сессия
	wg.Add(2)
	go dl.loadComments(&wg, commentsReady)
	go dl.loadSession(&wg)

	// После комментариев: пользователи
	wg.Add(1)
	go dl.loadUsers(&wg, commentsReady)

	// Условно: вложения (только если есть session-id)
	if dl.sessionID != "" {
		wg.Add(1)
		go dl.loadAttachments(&wg)
	}

	wg.Wait()
	fmt.Println("Все данные загружены синхронно")
}

// вывод результата
func (dl *DataLoader) PrintSummary() {
	fmt.Println("Итоги загрузки:")
	fmt.Printf("Комментарии: %d\n", len(dl.comments))
	fmt.Printf("Пользователи: %d\n", len(dl.users))
	if dl.session != nil {
		fmt.Printf("Сессия: %s (до %s)\n", dl.session.ID, dl.session.ExpiresAt.Format("15:04:05"))
	} else {
		fmt.Println("Сессия: не загружена")
	}
	fmt.Printf("Вложения: %d\n", len(dl.attachments))
}

func main() {
	fmt.Println("Параллельная загрузка данных\n")

	fmt.Println("Тест 1: С session-id")
	loader1 := NewDataLoader("sess_abc123")
	loader1.LoadAll()
	loader1.PrintSummary()

	fmt.Println("\n" + "----")

	fmt.Println("Тест 2:Без session-id")
	loader2 := NewDataLoader("")
	loader2.LoadAll()
	loader2.PrintSummary()
}
