package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// хранит результаты обработки одного файла
type Result struct {
	Filename  string
	WordCount int
	Error     error
}

// подсчитывает количество слов в одном файле
func countWordsIsFile(path string) Result {
	file, err := os.Open(path)
	if err != nil {
		return Result{Filename: path, Error: err}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)

	count := 0
	for scanner.Scan() {
		_ = scanner.Text()
		count++
	}

	if err := scanner.Err(); err != nil {
		return Result{Filename: path, Error: err}
	}

	return Result{Filename: filepath.Base(path), WordCount: count}

}

// распределяет файлы по горутинам и собирает результаты
func fanOut(files []string, workerCount int) <-chan Result {
	results := make(chan Result)
	var wg sync.WaitGroup

	jobs := make(chan string, len(files))

	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range jobs {
				results <- countWordsIsFile(file)
			}
		}()
	}

	//Отправляем все файлы в канал задач
	go func() {
		for _, f := range files {
			jobs <- f
		}
		close(jobs) // Сигнал воркерам: задач больше нет
	}()

	// Закрываем канал результатов, когда все воркеры завершатся
	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

func main() {
	// Путь к директории с файлами
	dir := "./testdata"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	//Собираем список файлов
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".txt") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Ошибка чтения директории: %v\n", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("Файлы .txt не найдены")
		return
	}

	// Запускаем Fan-Out с количеством воркеров = min(4, len(files))
	workerCount := 4
	if len(files) < workerCount {
		workerCount = len(files)
	}

	results := fanOut(files, workerCount)

	// Агрегируем результаты (Fan-In в main)
	totalWords := 0
	fileCount := 0
	errorCount := 0

	fmt.Println("\n Результаты по файлам:")
	fmt.Println(strings.Repeat("-", 40))
	for res := range results {
		if res.Error != nil {
			fmt.Printf("%s: %v\n", res.Filename, res.Error)
			errorCount++
		} else {
			fmt.Printf("%s: %d слов\n", res.Filename, res.WordCount)
			totalWords += res.WordCount
			fileCount++
		}
	}

	//Итоговая статистика
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("Обработано файлов: %d\n", fileCount)
	fmt.Printf("Ошибок: %d\n", errorCount)
	fmt.Printf("Всего слов: %d\n", totalWords)
	if fileCount > 0 {
		fmt.Printf("В среднем: %.1f слов/файл\n", float64(totalWords)/float64(fileCount))
	}
}
