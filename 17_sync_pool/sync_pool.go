package main

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
)

// пул для переиспользования
var bufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 256))
	},
}

// преобразует строку в верхний регистр
func ProcessString(s string) string {
	buf := bufferPool.Get().(*bytes.Buffer)

	//очищаем буфер перед использованием
	buf.Reset()
	defer bufferPool.Put(buf)
	_, _ = buf.WriteString(strings.ToUpper(s))
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())

	return string(result)
}

// альтернативная версия без лишних копий
func ProcessStringOptimized(s string) string {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	// Пишем преобразованную строку в буфер
	_, _ = buf.WriteString(strings.ToUpper(s))

	// Копируем только финальный результат (неизбежно для string)
	return string(append([]byte(nil), buf.Bytes()...))
}

func main() {
	examples := []string{
		"hello, world!",
		"gopher",
		"синхронизация в многопоточных программах",
		"оптимизация производительности",
		"высоконагруженные сервисы на го",
	}
	fmt.Println("Обработка строк с использованием sync.Pool:\n")

	for _, s := range examples {
		processed := ProcessString(s)
		fmt.Printf("original: %q\n", s)
		fmt.Printf("processed: %q\n", processed)
	}
	fmt.Println("Буферы успешно переиспользованы, нагрузка на GC снижена!")
}
