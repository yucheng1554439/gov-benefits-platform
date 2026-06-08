package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const queueKey = "govbenefits:jobs"

type JobType string

const (
	JobGenerateLetter JobType = "generate_letter"
	JobGenerateReport JobType = "generate_report"
	JobRunRetention   JobType = "run_retention"
)

type Job struct {
	ID      string         `json:"id"`
	Type    JobType        `json:"type"`
	Payload map[string]any `json:"payload"`
}

type Queue struct {
	client *redis.Client
}

func NewQueue(redisURL string) (*Queue, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return &Queue{client: client}, nil
}

func (q *Queue) Enqueue(ctx context.Context, job Job) error {
	if job.ID == "" {
		job.ID = uuid.New().String()
	}
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return q.client.LPush(ctx, queueKey, data).Err()
}

func (q *Queue) Dequeue(ctx context.Context, timeout time.Duration) (*Job, error) {
	result, err := q.client.BRPop(ctx, timeout, queueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	if len(result) < 2 {
		return nil, nil
	}
	var job Job
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func (q *Queue) Close() error {
	return q.client.Close()
}

func (q *Queue) Ping(ctx context.Context) error {
	return q.client.Ping(ctx).Err()
}
