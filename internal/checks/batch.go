package checks

import (
	"context"
	"sync"
)

type BatchItem[T any] struct {
	Index  int
	Item   string
	Result T
}

func RunBatch[T any](ctx context.Context, items []string, concurrency int, check func(context.Context, string) T) []BatchItem[T] {
	results := make([]BatchItem[T], len(items))
	if len(items) == 0 {
		return results
	}
	if concurrency <= 0 {
		concurrency = 4
	}
	if concurrency > len(items) {
		concurrency = len(items)
	}

	jobs := make(chan int, len(items))
	for i := range items {
		jobs <- i
	}
	close(jobs)

	var workers sync.WaitGroup
	workers.Add(concurrency)
	for range concurrency {
		go func() {
			defer workers.Done()
			for index := range jobs {
				results[index] = BatchItem[T]{Index: index, Item: items[index], Result: check(ctx, items[index])}
			}
		}()
	}
	workers.Wait()
	return results
}
