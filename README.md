# nats-bus

`nats-bus` - is a wrapper for NATS focussed on NATS JetStream.

**Key points**:
* Stream per topic(subject)
* Durable consumers
* Automatically stream name generating `topic_{topicName}`
* Automatically consumer name generating `{serviceName}_{streamName}_consumer`
* Automatically stream creating on consuming
* Automatically stream creating on producing

## Install

``
go get github.com/artarts36/nats-bus
``

## Example: event definition

```go
type TestEvent struct {
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
}

func (TestEvent) TopicName() string {
    return "users.created.v1"
}

func (TestEvent) CreateFromJSON([]byte) (Event, error) {
    var event TestEvent
    
    err := json.Unmarshal(val, &event)
    if err != nil {
        return nil, err
    }
    
    return event, nil
}
```

## Example: publish

```go
package main

import (
	"context"
	"log"

	"github.com/artarts36/nats-bus"
)

func main() {
	bus, err := natsbus.Connect(&natsbus.Config{
		ServiceName: "name_of_your_service",
		NatsURLs:    []string{"nats1:4222", "nats2:4222", "nats3:4222"},
	})
	if err != nil {
		log.Fatalln("failed to connect to nats", err.Error())
	}

	err = bus.Publish(context.Background(), &TestEvent{})
	if err != nil {
		log.Fatalln("failed to connect to nats", err.Error())
	}
}
```

## Example: consume

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/artarts36/nats-bus"
)

func main() {
    bus, err := natsbus.Connect(&natsbus.Config{
        ServiceName: "name_of_your_service",
        NatsURLs:    []string{"nats1:4222", "nats2:4222", "nats3:4222"},
    })
    if err != nil {
        log.Fatalln("failed to connect to nats", err.Error())
    }
    
    bus.Subscribe(&natsbus.EventSubscriber{
        Event: &TestEvent{},
        Subscriber: func(event *natsbus.ConsumedEvent) error {
            fmt.Println(event)
            return nil
        },
    })
    
    err := bus.Consume(context.Background())
    if err != nil {
        log.Fatalln("failed to connect to consume", err.Error())
    }
}
```
