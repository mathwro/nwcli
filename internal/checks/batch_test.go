package checks

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestRunBatchBoundsConcurrencyAndPreservesOrder(t *testing.T) {
	items := []string{"one", "two", "three", "four"}
	started := make(chan string, len(items))
	release := make(chan struct{})
	done := make(chan []BatchItem[string], 1)
	go func() {
		done <- RunBatch(context.Background(), items, 2, func(_ context.Context, item string) string {
			started <- item
			<-release
			return "checked-" + item
		})
	}()

	<-started
	<-started
	select {
	case item := <-started:
		t.Fatalf("third item %q started before a worker was released", item)
	case <-time.After(20 * time.Millisecond):
	}
	close(release)
	results := <-done
	values := make([]string, len(results))
	for i, result := range results {
		if result.Item != items[i] || result.Index != i {
			t.Fatalf("result[%d] = %#v", i, result)
		}
		values[i] = result.Result
	}
	want := []string{"checked-one", "checked-two", "checked-three", "checked-four"}
	if !reflect.DeepEqual(values, want) {
		t.Fatalf("values = %#v, want %#v", values, want)
	}
}

func TestRunBatchUsesSafeDefaultConcurrency(t *testing.T) {
	results := RunBatch(context.Background(), []string{"one"}, 0, func(_ context.Context, item string) string { return item })
	if len(results) != 1 || results[0].Result != "one" {
		t.Fatalf("results = %#v", results)
	}
}
