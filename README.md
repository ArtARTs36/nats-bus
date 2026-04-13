# nats-bus

`nats-bus` - wrapper for NATS focussed on NATS JetStream.

**Key points**:
* Stream per topic(subject)
* Durable consumers
* Automatically stream name generating `topic_{topicName}`
* Automatically consumer name generating `{serviceName}_{streamName}_consumer`
* Automatically stream creating on consuming
* Automatically stream creating on producing
* Pluggable `Serializer`: built-in `NewJSONSerializer` and `NewProtoSerializer`, or custom via `NewFuncSerializer`; `ComposeWithJSON` / `NewComposeSerializer` route by `SerializationType`

## Install

```bash
go get github.com/artarts36/nats-bus
```

## Example: event definition

```go
type TestEvent struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (e *TestEvent) TopicMeta() natsbus.TopicMeta {
	return natsbus.TopicMeta{
		TopicName:         "users.created.v1",
		SerializationType: natsbus.SerializationJSON,
	}
}
```

Use `natsbus.NewComposeSerializer` when you register more than one `SerializationType`, or `natsbus.ComposeWithJSON(extra)` to keep JSON and add your own `SerializationType` values.

## Custom serializer

`Connect` accepts any `natsbus.Serializer`. Options:

1. **Struct** — implement `Serialize` / `Deserialize` on your type.
2. **`NewFuncSerializer`** — pass two functions (same contract as the interface).
3. **Composition** — `ComposeWithJSON(map[natsbus.SerializationType]natsbus.Serializer{"my-format": mySer})` registers JSON plus custom kinds; omit the map or pass `nil` to only get the default JSON serializer.

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
	}, natsbus.NewJSONSerializer())
	if err != nil {
		log.Fatalln("failed to connect to nats", err.Error())
	}

	_, err = bus.Publish(context.Background(), &TestEvent{})
	if err != nil {
		log.Fatalln("failed to publish", err.Error())
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
	}, natsbus.NewJSONSerializer())
	if err != nil {
		log.Fatalln("failed to connect to nats", err.Error())
	}

	bus.Subscribe(&natsbus.EventSubscriber{
		Event: &TestEvent{},
		Subscriber: func(ctx context.Context, event *natsbus.ConsumedEvent) error {
			fmt.Println(event)
			return nil
		},
	})

	err = bus.Consume(context.Background())
	if err != nil {
		log.Fatalln("failed to consume", err.Error())
	}
}
```
