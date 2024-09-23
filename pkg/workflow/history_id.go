package workflow

import natsbus "github.com/artarts36/nats-bus"

func HistoryID() func(event natsbus.Event) (string, bool) {
	return func(busEvent natsbus.Event) (string, bool) {
		ev, ok := busEvent.(*event)
		if !ok {
			return "", false
		}
		return ev.WorkflowID, false
	}
}
