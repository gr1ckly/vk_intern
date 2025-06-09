package workerpool

import "fmt"

type WorkerRestartError struct {
	id int
}

func (wre WorkerRestartError) Error() string {
	return fmt.Sprintf("Failing to restart worker with id: %v", wre.id)
}

func NewWorkerRestartError(id int) WorkerRestartError {
	return WorkerRestartError{id}
}

type WorkerError struct {
	id  int
	err error
}

func (we WorkerError) Error() string {
	return fmt.Sprintf("Worker %v error: %v", we.id, we.err)
}

func (we WorkerError) Unwrap() error {
	return we.err
}

type AlreadyStartedError struct {
	id int
}

func (ase AlreadyStartedError) Error() string {
	return fmt.Sprintf("Worker %v has already started", ase.id)
}

func NewAlreadyStartedError(id int) AlreadyStartedError {
	return AlreadyStartedError{id}
}

type DoesntExistsWorkerError struct {
	id int
}

func (dewe DoesntExistsWorkerError) Error() string {
	return fmt.Sprintf("Worker %v doesn't exists", dewe.id)
}

func NewDoesntExistsWorkerError(id int) DoesntExistsWorkerError {
	return DoesntExistsWorkerError{id}
}
