package workflow

import (
	"context"
	"fmt"
	natsbus "github.com/artarts36/nats-bus"
)

type Manager struct {
	bus natsbus.Bus

	workflows    map[string]*Workflow
	activityInfo map[string]*ActivityInfo
}

func NewManager(bus natsbus.Bus) *Manager {
	return &Manager{
		bus:       bus,
		workflows: map[string]*Workflow{},
	}
}

func (b *Manager) Register(wf *Workflow) {
	b.workflows[wf.Name] = wf

	wf.Register(b.bus)

	if len(b.activityInfo) == 0 {
		b.activityInfo = wf.activityInfo

		return
	}

	for k, v := range wf.activityInfo {
		b.activityInfo[k] = v
	}
}

func (b *Manager) Dispatch(ctx context.Context, workflowName string, payload []byte) error {
	wf, exists := b.workflows[workflowName]
	if !exists {
		return fmt.Errorf("workflow with name %q not found", workflowName)
	}

	return wf.pushNewWorkflow(ctx, b.bus, payload)
}

func (b *Manager) Consume(ctx context.Context) error {
	return b.bus.Consume(ctx)
}

func (b *Manager) GetActivityInfo() map[string]*ActivityInfo {
	return b.activityInfo
}
