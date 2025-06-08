package worker

import (
	"fmt"
	"log"
	"poetry/db"
	"time"
)

type Worker struct {
	connection *db.MongoDBConnection
	jobChan    chan Job
	quit       chan bool
	maxWorkers int
}

type Job struct {
	Poems []db.Poem
}

// NewWorker creates a new worker instance
func NewWorker(connection *db.MongoDBConnection, bufferSize int, maxWorkers int) *Worker {
	return &Worker{
		connection: connection,
		jobChan:    make(chan Job, bufferSize),
		quit:       make(chan bool),
		maxWorkers: maxWorkers,
	}
}

// Start begins the worker processing jobs
func (w *Worker) Start() {
	log.Printf("Starting worker with %d goroutines", w.maxWorkers)

	for i := 0; i < w.maxWorkers; i++ {
		go w.processJobs(i)
	}
}

// Stop gracefully shuts down the worker
func (w *Worker) Stop() {
	log.Println("Stopping worker...")
	close(w.quit)
}

// AddJob adds a new job to the queue
func (w *Worker) AddJob(poems []db.Poem) error {
	job := Job{Poems: poems}

	select {
	case w.jobChan <- job:
		log.Printf("Job added with %d poems", len(poems))
		return nil
	default:
		return fmt.Errorf("worker queue is full, job rejected")
	}
}

// GetQueueSize returns the current number of jobs in the queue
func (w *Worker) GetQueueSize() int {
	return len(w.jobChan)
}

// processJobs processes jobs from the job channel
func (w *Worker) processJobs(workerID int) {
	log.Printf("Worker %d started", workerID)

	for {
		select {
		case job := <-w.jobChan:
			w.processJob(workerID, job)
		case <-w.quit:
			log.Printf("Worker %d stopping", workerID)
			return
		}
	}
}

// processJob processes a single job
func (w *Worker) processJob(workerID int, job Job) {
	start := time.Now()
	log.Printf("Worker %d processing job with %d poems", workerID, len(job.Poems))

	// Convert poems to documents
	var documents []interface{}
	for _, poem := range job.Poems {
		documents = append(documents, poem)
	}

	// Get collection and insert documents
	collection, err := db.GetCollection("poetry", "poems", w.connection)
	if err != nil {
		log.Printf("Worker %d error getting collection: %v", workerID, err)
		return
	}

	db.InsertManyIntoDB(*collection, documents)

	duration := time.Since(start)
	log.Printf("Worker %d completed job in %v", workerID, duration)
}
