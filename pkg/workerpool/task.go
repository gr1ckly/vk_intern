package workerpool

type Task[T any] interface {
	Do(T) error
}
