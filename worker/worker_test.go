package worker

import (
	"fmt"
	"poetry/db"
	"sync"
	"testing"
	"time"
)

func TestNewWorker(t *testing.T) {
	conn := &db.MongoDBConnection{}
	bufferSize := 5
	maxWorkers := 2

	worker := NewWorker(conn, bufferSize, maxWorkers)

	if worker == nil {
		t.Fatal("NewWorker returned nil")
	}

	if worker.connection != conn {
		t.Error("Worker connection not set correctly")
	}

	if cap(worker.jobChan) != bufferSize {
		t.Errorf("Expected job channel buffer size %d, got %d", bufferSize, cap(worker.jobChan))
	}

	if worker.maxWorkers != maxWorkers {
		t.Errorf("Expected max workers %d, got %d", maxWorkers, worker.maxWorkers)
	}
}

func TestWorkerAddJob_Success(t *testing.T) {
	worker := NewWorker(&db.MongoDBConnection{}, 2, 1)

	poems := []db.Poem{
		{Title: "Test Poem 1", Poem: "This is a test poem", Language: "en"},
		{Title: "Test Poem 2", Poem: "Another test poem", Language: "en"},
	}

	err := worker.AddJob(poems)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if worker.GetQueueSize() != 1 {
		t.Errorf("Expected queue size 1, got %d", worker.GetQueueSize())
	}
}

func TestWorkerAddJob_QueueFull(t *testing.T) {
	worker := NewWorker(&db.MongoDBConnection{}, 1, 1)

	poems1 := []db.Poem{{Title: "Poem 1", Language: "en"}}
	poems2 := []db.Poem{{Title: "Poem 2", Language: "en"}}

	// Fill the queue
	err1 := worker.AddJob(poems1)
	if err1 != nil {
		t.Errorf("First job should succeed, got error: %v", err1)
	}

	// This should fail because queue is full
	err2 := worker.AddJob(poems2)
	if err2 == nil {
		t.Error("Expected error when queue is full, got nil")
	}

	if err2.Error() != "worker queue is full, job rejected" {
		t.Errorf("Expected specific error message, got: %v", err2)
	}
}

func TestWorkerGetQueueSize(t *testing.T) {
	worker := NewWorker(&db.MongoDBConnection{}, 3, 1)

	if worker.GetQueueSize() != 0 {
		t.Errorf("Expected initial queue size 0, got %d", worker.GetQueueSize())
	}

	poems := []db.Poem{{Title: "Test", Language: "en"}}
	worker.AddJob(poems)

	if worker.GetQueueSize() != 1 {
		t.Errorf("Expected queue size 1 after adding job, got %d", worker.GetQueueSize())
	}
}

func TestWorkerStartAndStop(t *testing.T) {
	worker := NewWorker(&db.MongoDBConnection{}, 2, 2)

	// Start the worker
	worker.Start()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Stop the worker
	worker.Stop()

	// Test passes if no panic occurs
}

func TestWorkerConcurrency(t *testing.T) {
	worker := NewWorker(&db.MongoDBConnection{}, 10, 3)
	// Don't start worker to avoid database operations

	var wg sync.WaitGroup
	numJobs := 5

	for i := 0; i < numJobs; i++ {
		wg.Add(1)
		go func(jobNum int) {
			defer wg.Done()
			poems := []db.Poem{
				{Title: fmt.Sprintf("Concurrent Poem %d", jobNum), Language: "en"},
			}
			err := worker.AddJob(poems)
			if err != nil {
				t.Errorf("Job %d failed: %v", jobNum, err)
			}
		}(i)
	}

	wg.Wait()
	// Test queue filling behavior
	if worker.GetQueueSize() != numJobs {
		t.Errorf("Expected queue size %d, got %d", numJobs, worker.GetQueueSize())
	}
}

func TestWorkerMultipleWorkers(t *testing.T) {
	// Test with multiple worker goroutines - don't start to avoid DB operations
	worker := NewWorker(&db.MongoDBConnection{}, 20, 5)

	// Add multiple jobs quickly
	for i := 0; i < 10; i++ {
		poems := []db.Poem{
			{Title: fmt.Sprintf("Multi-worker Poem %d", i), Language: "en"},
		}
		err := worker.AddJob(poems)
		if err != nil {
			t.Errorf("Job %d failed: %v", i, err)
		}
	}

	// Verify jobs were queued
	if worker.GetQueueSize() != 10 {
		t.Errorf("Expected queue size 10, got %d", worker.GetQueueSize())
	}
}

func TestWorkerLargeJob(t *testing.T) {
	worker := NewWorker(&db.MongoDBConnection{}, 5, 2)

	// Create a large batch of poems
	var poems []db.Poem
	for i := 0; i < 100; i++ {
		poems = append(poems, db.Poem{
			Title:    fmt.Sprintf("Large Job Poem %d", i),
			Poem:     fmt.Sprintf("Content for poem number %d", i),
			Language: "en",
			Poet:     "Test Author",
		})
	}

	err := worker.AddJob(poems)
	if err != nil {
		t.Errorf("Large job failed: %v", err)
	}

	if worker.GetQueueSize() != 1 {
		t.Errorf("Expected queue size 1 after adding large job, got %d", worker.GetQueueSize())
	}
}

func TestWorkerQueueOverload(t *testing.T) {
	// Test behavior when worker is overloaded
	worker := NewWorker(&db.MongoDBConnection{}, 2, 1)

	// Fill the queue
	for i := 0; i < 2; i++ {
		poems := []db.Poem{{Title: fmt.Sprintf("Poem %d", i), Language: "en"}}
		err := worker.AddJob(poems)
		if err != nil {
			t.Errorf("Job %d should succeed, got error: %v", i, err)
		}
	}

	// This should fail because queue is full
	overflowPoems := []db.Poem{{Title: "Overflow Poem", Language: "en"}}
	err := worker.AddJob(overflowPoems)
	if err == nil {
		t.Error("Expected error when queue is overloaded, got nil")
	}

	expectedError := "worker queue is full, job rejected"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestWorkerEmptyJob(t *testing.T) {
	worker := NewWorker(&db.MongoDBConnection{}, 5, 1)

	// Test with empty poems slice
	err := worker.AddJob([]db.Poem{})
	if err != nil {
		t.Errorf("Adding empty job should succeed, got error: %v", err)
	}

	if worker.GetQueueSize() != 1 {
		t.Errorf("Expected queue size 1 after adding empty job, got %d", worker.GetQueueSize())
	}
}

func TestWorkerJobStructure(t *testing.T) {
	worker := NewWorker(&db.MongoDBConnection{}, 5, 1)

	poems := []db.Poem{
		{
			Title:     "Test Poem",
			Poem:      "This is a test poem content",
			Language:  "en",
			Poet:      "Test Author",
			Dataset:   "test_dataset",
			DatasetId: "test_001",
			Tags:      []string{"test", "poem"},
		},
	}

	err := worker.AddJob(poems)
	if err != nil {
		t.Errorf("Adding structured job should succeed, got error: %v", err)
	}

	if worker.GetQueueSize() != 1 {
		t.Errorf("Expected queue size 1 after adding job, got %d", worker.GetQueueSize())
	}
}

func TestWorkerStressTest(t *testing.T) {
	// Stress test with many concurrent operations - don't start to avoid DB operations
	worker := NewWorker(&db.MongoDBConnection{}, 50, 5)

	var wg sync.WaitGroup
	numGoroutines := 20
	jobsPerGoroutine := 5
	var successCount int
	var failureCount int
	var mutex sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < jobsPerGoroutine; j++ {
				poems := []db.Poem{
					{
						Title:    fmt.Sprintf("Stress Test Poem G%d-J%d", goroutineID, j),
						Language: "en",
						Poet:     fmt.Sprintf("Author %d", goroutineID),
					},
				}
				err := worker.AddJob(poems)
				mutex.Lock()
				if err != nil {
					failureCount++
				} else {
					successCount++
				}
				mutex.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Verify that we had both successes and failures (demonstrating overload protection)
	totalJobs := numGoroutines * jobsPerGoroutine
	if successCount+failureCount != totalJobs {
		t.Errorf("Expected %d total job attempts, got %d", totalJobs, successCount+failureCount)
	}

	// Should have successes up to buffer capacity
	if successCount > 50 {
		t.Errorf("Expected at most 50 successful jobs, got %d", successCount)
	}

	// Should have some failures due to overload protection
	if failureCount == 0 && totalJobs > 50 {
		t.Error("Expected some failures due to queue overload, but got none")
	}

	// Queue size should match successful submissions
	if worker.GetQueueSize() != successCount {
		t.Errorf("Expected queue size %d, got %d", successCount, worker.GetQueueSize())
	}
}
