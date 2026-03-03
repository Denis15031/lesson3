package main

import (
	"testing"
)

func TestMergeChannels(t *testing.T) {
	ch1 := make(chan int)
	ch2 := make(chan int)

	go func() {
		close(ch1)
		close(ch2)
	}()

	merged := mergeChannels(ch1, ch2)

	// Канал должен закрыться
	for range merged {
		t.Error("Ожидалось закрытие канала")
	}
}
