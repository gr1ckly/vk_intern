package workerpool_test

import (
	"VK_intern_test/pkg/workerpool"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
)

type StringTask struct{}

func (st StringTask) Do(line string) error {
	fmt.Println(line)
	return nil
}

type ErrorTask struct{}

func (et ErrorTask) Do(line string) error {
	return errors.New("simulated task error")
}

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String()
}

func TestWorkerPoolStringTaskBasic(t *testing.T) {
	dataChan := make(chan string)
	errChan := make(chan error, 10)
	pool := workerpool.NewWorkerPool[string](StringTask{}, context.Background(), dataChan, errChan)
	defer pool.Stop()
	for i := 0; i < 3; i++ {
		if _, err := pool.AddWorker(); err != nil {
			t.Fatalf("AddWorker failed: %v", err)
		}
	}
	output := captureOutput(func() {
		go func() {
			defer close(dataChan)
			for _, msg := range []string{"one", "two", "three"} {
				dataChan <- msg
			}
		}()
		pool.Wait()
	})
	for _, msg := range []string{"one", "two", "three"} {
		if !strings.Contains(output, msg) {
			t.Errorf("expected output to contain %q", msg)
		}
	}
}

func TestWorkerPoolRemoveWorker(t *testing.T) {
	dataChan := make(chan string)
	errChan := make(chan error, 10)
	pool := workerpool.NewWorkerPool[string](StringTask{}, context.Background(), dataChan, errChan)
	defer pool.Stop()
	id, err := pool.AddWorker()
	if err != nil {
		t.Fatalf("AddWorker failed: %v", err)
	}
	if err = pool.RemoveWorker(id); err != nil {
		t.Errorf("RemoveWorker failed: %v", err)
	}
	if err = pool.RemoveWorker(id); err == nil {
		t.Error("expected error on removing already removed worker")
	} else if _, ok := err.(workerpool.DoesntExistsWorkerError); !ok {
		t.Errorf("expected DoesntExistsWorkerError, got %v", err)
	}
}

func TestWorkerPoolErrorHandling(t *testing.T) {
	dataChan := make(chan string)
	errChan := make(chan error, 1)
	pool := workerpool.NewWorkerPool[string](ErrorTask{}, context.Background(), dataChan, errChan)
	defer pool.Stop()
	if _, err := pool.AddWorker(); err != nil {
		t.Fatalf("AddWorker failed: %v", err)
	}
	go func() {
		dataChan <- "fail"
		close(dataChan)
	}()
	pool.Wait()
	select {
	case err := <-errChan:
		if err == nil || !strings.Contains(err.Error(), "simulated") {
			t.Errorf("unexpected error: %v", err)
		}
	default:
		t.Error("expected error but got none")
	}
}

func TestWorkerPoolStop(t *testing.T) {
	dataChan := make(chan string)
	errChan := make(chan error)
	pool := workerpool.NewWorkerPool[string](StringTask{}, context.Background(), dataChan, errChan)
	defer pool.Stop()
	for i := 0; i < 5; i++ {
		if _, err := pool.AddWorker(); err != nil {
			t.Fatalf("AddWorker failed: %v", err)
		}
	}
}

func TestWorkerPoolGoroutineLeak(t *testing.T) {
	before := runtime.NumGoroutine()
	dataChan := make(chan string)
	errChan := make(chan error)
	pool := workerpool.NewWorkerPool[string](StringTask{}, context.Background(), dataChan, errChan)
	for i := 0; i < 3; i++ {
		if _, err := pool.AddWorker(); err != nil {
			t.Fatalf("AddWorker failed: %v", err)
		}
	}
	close(dataChan)
	pool.Wait()
	after := runtime.NumGoroutine()
	if diff := after - before; diff > 2 {
		t.Errorf("potential goroutine leak: before=%d after=%d", before, after)
	}
}

func TestWorkerPoolAddWorkerConcurrent(t *testing.T) {
	dataChan := make(chan string)
	errChan := make(chan error)
	pool := workerpool.NewWorkerPool[string](StringTask{}, context.Background(), dataChan, errChan)
	defer pool.Stop()
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := pool.AddWorker(); err != nil {
				t.Errorf("AddWorker failed: %v", err)
			}
		}()
	}
	wg.Wait()
	close(dataChan)
	pool.Wait()
}
