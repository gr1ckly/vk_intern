package main

import (
	"VK_intern_test/pkg/workerpool"
	"fmt"
	"strconv"
)

type StringTask struct{}

func (st StringTask) Do(line string) error {
	fmt.Println(line)
	return nil
}

func main() {
	errChan := make(chan error)
	dataChan := make(chan string)
	wp := workerpool.NewWorkerPool[string](StringTask{}, dataChan, errChan)
	defer wp.Stop()
	go func() {
		defer close(dataChan)
		for err := range errChan {
			fmt.Printf("Error: %v", err)
		}
	}()
	for range 10 {
		wp.AddWorker()
	}
	for i := 0; i < 100; i++ {
		dataChan <- strconv.Itoa(i)
	}
	close(dataChan)
	wp.Wait()
}
