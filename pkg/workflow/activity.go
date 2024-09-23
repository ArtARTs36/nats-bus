package workflow

import "context"

type ActivityTask struct {
	Payload interface{}
}

type Activity struct {
	Name        string
	Action      func(ctx context.Context, task *ActivityTask) (interface{}, error)
	MaxAttempts int

	topicName string
}

type ActivityInfo struct {
	Name        string
	MaxAttempts int
	TopicName   string
}
