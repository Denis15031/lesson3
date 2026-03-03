package main

import (
	"context"
	"fmt"
	"time"
)

// метрика сервиса(значение в юайтах)
type ServiceMetric struct {
	Name  string
	Value float64
}

// конвертирует байты в мегабайты
const bytesToMB = 1024 * 1024

// читает из входного, преобразует value из Б в МБ, отправляет в выходной
func TransformMetrics(ctx context.Context, in <-chan ServiceMetric) <-chan ServiceMetric {
	out := make(chan ServiceMetric)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case metric, ok := <-in:
				if !ok {
					return
				}
				metric.Value = metric.Value / bytesToMB
				out <- metric
			}
		}
	}()
	return out
}

// генератор тестовых метрик
func generateMetrics(ctx context.Context) <-chan ServiceMetric {
	out := make(chan ServiceMetric)

	go func() {
		defer close(out)
		metrics := []ServiceMetric{
			{"memory_usage", 2_147_483_648},
			{"disk_io", 536_870_912},
			{"network_in", 1_073_741_824},
			{"cache_size", 268_435_456},
		}
		for _, m := range metrics {
			select {
			case <-ctx.Done():
				return
			case out <- m:
				time.Sleep(50 * time.Millisecond)
			}
		}
	}()
	return out
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rawMetrics := generateMetrics(ctx)

	mbMetrcics := TransformMetrics(ctx, rawMetrics)

	fmt.Println("Метрики в мегабайтах:")
	fmt.Println("-------------")
	for metric := range mbMetrcics {
		fmt.Printf("%s: %.2f MB\n", metric.Name, metric.Value)
	}
	fmt.Println("---------")
	fmt.Println("Обработка завершена")

}
