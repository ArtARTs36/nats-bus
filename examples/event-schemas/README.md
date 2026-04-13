# Example: event schemas

This folder shows how to use `protoc-gen-event-schemas` to:

1. generate Go event structs with the `TopicMeta()` method;
2. generate a markdown topic registry.

## Structure

1. `proto/events/venue_events.proto` - input event definitions.
2. `templates/topics.md.tmpl` - markdown documentation template.
3. `docs/topics.md` - generated topic registry.

## Generation

```bash
protoc \
  -I proto/third_party \
  --proto_path=. \
  --proto_path=./proto \
  --event-schemas_out=. \
  --event-schemas_opt=paths=source_relative,doc_template=./templates/topics.md.tmpl,doc_out=./docs/topics.md \
  ./proto/events/venue_events.proto
```

Generated files:

1. `proto/events/venue_events.event-schemas.pb.go`
2. `docs/topics.md`
