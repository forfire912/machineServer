package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/forfire912/machineServer/internal/model"
	"github.com/go-redis/redis/v8"
)

// Queue manages async job processing
type Queue struct {
	client *redis.Client
	ctx    context.Context
}

// NewQueue creates a new job queue
func NewQueue(redisAddr string, password string, db int) (*Queue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Queue{
		client: client,
		ctx:    ctx,
	}, nil
}

// EnqueueJob adds a job to the queue
func (q *Queue) EnqueueJob(job *model.Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	queueName := fmt.Sprintf("queue:%s", job.Type)
	return q.client.RPush(q.ctx, queueName, data).Err()
}

// DequeueJob retrieves a job from the queue
func (q *Queue) DequeueJob(jobType string, timeout time.Duration) (*model.Job, error) {
	queueName := fmt.Sprintf("queue:%s", jobType)
	
	result, err := q.client.BLPop(q.ctx, timeout, queueName).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No job available
		}
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	if len(result) < 2 {
		return nil, nil
	}

	var job model.Job
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return &job, nil
}

// GetQueueSize returns the number of jobs in a queue
func (q *Queue) GetQueueSize(jobType string) (int64, error) {
	queueName := fmt.Sprintf("queue:%s", jobType)
	return q.client.LLen(q.ctx, queueName).Result()
}

// UpdateJobStatus updates the status of a job
func (q *Queue) UpdateJobStatus(jobID string, status string, progress int) error {
	key := fmt.Sprintf("job:%s:status", jobID)
	data := map[string]interface{}{
		"status":   status,
		"progress": progress,
		"updated":  time.Now().Unix(),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return q.client.Set(q.ctx, key, jsonData, 24*time.Hour).Err()
}

// GetJobStatus retrieves the status of a job
func (q *Queue) GetJobStatus(jobID string) (map[string]interface{}, error) {
	key := fmt.Sprintf("job:%s:status", jobID)
	data, err := q.client.Get(q.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var status map[string]interface{}
	if err := json.Unmarshal([]byte(data), &status); err != nil {
		return nil, err
	}

	return status, nil
}

// Close closes the queue connection
func (q *Queue) Close() error {
	return q.client.Close()
}

// Worker processes jobs from the queue
type Worker struct {
	queue     *Queue
	jobType   string
	handler   JobHandler
	stopChan  chan bool
}

// JobHandler is a function that processes a job
type JobHandler func(job *model.Job) error

// NewWorker creates a new job worker
func NewWorker(queue *Queue, jobType string, handler JobHandler) *Worker {
	return &Worker{
		queue:    queue,
		jobType:  jobType,
		handler:  handler,
		stopChan: make(chan bool),
	}
}

// Start starts the worker
func (w *Worker) Start() {
	go func() {
		for {
			select {
			case <-w.stopChan:
				return
			default:
				job, err := w.queue.DequeueJob(w.jobType, 5*time.Second)
				if err != nil {
					continue
				}

				if job != nil {
					w.queue.UpdateJobStatus(job.ID, "running", 0)
					
					if err := w.handler(job); err != nil {
						w.queue.UpdateJobStatus(job.ID, "failed", 0)
					} else {
						w.queue.UpdateJobStatus(job.ID, "completed", 100)
					}
				}
			}
		}
	}()
}

// Stop stops the worker
func (w *Worker) Stop() {
	w.stopChan <- true
}
