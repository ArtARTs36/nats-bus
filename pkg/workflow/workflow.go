package workflow

import (
	"context"
	"fmt"
	"github.com/google/uuid"

	natsbus "github.com/artarts36/nats-bus"
)

type Workflow struct {
	Name       string
	Activities []*Activity

	activityInfo map[string]*ActivityInfo
}

type activitySubscriber struct {
	activity        *Activity
	eventSubscriber *natsbus.EventSubscriber
}

func (w *Workflow) Register(bus natsbus.Bus) {
	for i, activity := range w.Activities {
		var nextActivity *Activity
		if i+1 < len(w.Activities)-1 {
			nextActivity = w.Activities[i+1]
		}

		activity.topicName = fmt.Sprintf("%s_%s", w.Name, activity.Name)
		w.activityInfo[activity.topicName] = &ActivityInfo{
			Name:        w.Name,
			MaxAttempts: activity.MaxAttempts,
			TopicName:   activity.topicName,
		}

		subscriber := w.createEventSubscriber(bus, activity, nextActivity)

		bus.Subscribe(subscriber.eventSubscriber)
	}
}

func (w *Workflow) pushNewWorkflow(ctx context.Context, bus natsbus.Bus, payload interface{}) error {
	if len(w.Activities) == 0 {
		return fmt.Errorf("workflow not have activities")
	}

	activity := w.Activities[0]

	_, err := bus.Publish(ctx, &event{
		WorkflowID:   uuid.NewString(),
		WorkflowName: w.Name,
		ActivityName: activity.Name,
		Payload:      payload,
		topicName:    activity.topicName,
	})

	return err
}

func (w *Workflow) createEventSubscriber(bus natsbus.Bus, activity *Activity, nextActivity *Activity) *activitySubscriber {
	actSubscriber := &activitySubscriber{
		activity: activity,
	}

	actSubscriber.eventSubscriber = &natsbus.EventSubscriber{
		Event: &event{
			WorkflowName: w.Name,
			ActivityName: activity.Name,
		},
		Subscriber: func(ctx context.Context, ev *natsbus.ConsumedEvent) error {
			wfEvent, ok := ev.Event.(*event)
			if !ok {
				return fmt.Errorf("failed to parse workflow event")
			}

			result, err := activity.Action(ctx, &ActivityTask{Payload: wfEvent.Payload})
			if err != nil {
				return fmt.Errorf("failed to run activity action: %w", err)
			}

			if nextActivity != nil {
				_, err = bus.Publish(ctx, &event{
					WorkflowID:   wfEvent.WorkflowID,
					WorkflowName: w.Name,
					ActivityName: nextActivity.Name,
					Payload:      result,
					topicName:    activity.topicName,
				})
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	return actSubscriber
}
