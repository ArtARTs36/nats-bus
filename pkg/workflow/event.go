package workflow

import (
	"encoding/json"
	natsbus "github.com/artarts36/nats-bus"
)

type event struct {
	WorkflowID   string      `json:"workflow_id"`
	WorkflowName string      `json:"workflow_name"`
	ActivityName string      `json:"activity_name"`
	Payload      interface{} `json:"payload"`

	topicName string
}

func (e *event) TopicName() string {
	return e.topicName
}

func (e *event) CreateFromJSON(payload []byte) (natsbus.Event, error) {
	var ev event

	err := json.Unmarshal(payload, &ev)
	if err != nil {
		return nil, err
	}

	return &ev, nil
}
