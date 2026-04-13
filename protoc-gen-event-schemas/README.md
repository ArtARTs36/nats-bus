# protoc-gen-event-schemas

`protoc-gen-event-schemas` is a `protoc` plugin for generating Go event structures and a topic registry markdown document.

## What it generates

For each `message`:

1. Go struct with fields from `.proto`.
2. Topic constant.
3. `TopicMeta() natsbus.TopicMeta` method with:
   1. `TopicName` from `nats_bus.events.topic_name`.
   2. `SerializationType` from `nats_bus.events.serialization_type`.

It also generates one markdown file with a topics table using a custom template.

## Required message options

Use `proto/nats_bus/events/options.proto`:

```proto
import "nats_bus/events/options.proto";

// Event about user registered.
message UserRegistered {
  option (.nats_bus.events.topic_name) = "iam.users.registered";
  option (.nats_bus.events.serialization_type) = "json";

  string id = 1;
  string first_name = 2;
}
```

Known serialization values mapped to nats-bus constants:

1. `json` -> `natsbus.SerializationJSON`
2. `proto` / `protobuf` -> `natsbus.SerializationProto`

If `nats_bus.events.topic_name` or `nats_bus.events.serialization_type` is missing for a message, plugin generation fails with an error.

## Build plugin

Run from the `nats-bus` root:

```bash
go build -o ./bin/protoc-gen-event-schemas ./protoc-gen-event-schemas
```

## Run plugin

```bash
protoc \
  --proto_path=./proto \
  --proto_path=./protoc-gen-event-schemas/proto \
  --event-schemas_out=. \
  --event-schemas_opt=doc_template=./protoc-gen-event-schemas/templates/topics.md.tmpl,doc_out=./docs/topics.md \
  ./proto/events/user_events.proto
```

Plugin parameters:

1. `doc_template` (`doc-template` is also supported): path to markdown template.
2. `doc_out` (`doc-output` is also supported): output markdown path (relative to `--event-schemas_out`).
