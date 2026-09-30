package job

import (
	"log"
	"sync"
)

type Notification struct {
	TaskID   uint
	AuthorID uint
	Body     string
}

type Worker struct {
	queue chan Notification
	stop  chan struct{}
	group sync.WaitGroup
}

func NewWorker(buffer int) *Worker {
	worker := &Worker{queue: make(chan Notification, buffer), stop: make(chan struct{})}
	worker.group.Add(1)
	go worker.run()
	return worker
}

func (w *Worker) Enqueue(notification Notification) {
	select {
	case w.queue <- notification:
	default:
		log.Printf("notification queue full; dropping task %d notification", notification.TaskID)
	}
}

func (w *Worker) run() {
	defer w.group.Done()
	for {
		select {
		case notification := <-w.queue:
			log.Printf("background notification: user %d commented on task %d", notification.AuthorID, notification.TaskID)
		case <-w.stop:
			return
		}
	}
}

func (w *Worker) Close() {
	close(w.stop)
	w.group.Wait()
}
