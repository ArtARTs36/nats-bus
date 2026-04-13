package generator

import (
	"testing"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestSerializationTypeConstantGoName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "json",
			input: "json",
			want:  "SerializationJSON",
		},
		{
			name:  "proto",
			input: "proto",
			want:  "SerializationProto",
		},
		{
			name:  "protobuf alias",
			input: "protobuf",
			want:  "SerializationProto",
		},
		{
			name:  "unknown",
			input: "custom",
			want:  "",
		},
	}

	for _, testCase := range tests {
		tc := testCase

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := serializationTypeConstantGoName(tc.input)
			if got != tc.want {
				t.Fatalf("unexpected constant name: got %q want %q", got, tc.want)
			}
		})
	}
}

func TestParseEventOptions(t *testing.T) {
	raw := make([]byte, 0)
	raw = protowire.AppendTag(raw, topicNameOptionNumber, protowire.BytesType)
	raw = protowire.AppendString(raw, "iam.users.registered")
	raw = protowire.AppendTag(raw, serializationTypeOptionNumber, protowire.BytesType)
	raw = protowire.AppendString(raw, "json")

	options, err := parseEventOptions(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if options.TopicName != "iam.users.registered" {
		t.Fatalf("topic mismatch: got %q", options.TopicName)
	}

	if options.SerializationType != "json" {
		t.Fatalf("serialization mismatch: got %q", options.SerializationType)
	}
}

func TestParseEventOptions_InvalidWireType(t *testing.T) {
	raw := make([]byte, 0)
	raw = protowire.AppendTag(raw, topicNameOptionNumber, protowire.VarintType)
	raw = protowire.AppendVarint(raw, 1)

	_, err := parseEventOptions(raw)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestValidateEventOptions(t *testing.T) {
	messageName := protoreflect.FullName("events.UserRegistered")

	err := validateEventOptions(eventOptions{TopicName: "", SerializationType: "json"}, messageName)
	if err == nil {
		t.Fatal("expected topic_name validation error, got nil")
	}

	err = validateEventOptions(eventOptions{TopicName: "iam.users.registered", SerializationType: ""}, messageName)
	if err == nil {
		t.Fatal("expected serialization_type validation error, got nil")
	}

	err = validateEventOptions(eventOptions{
		TopicName:         "iam.users.registered",
		SerializationType: "json",
	}, messageName)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
