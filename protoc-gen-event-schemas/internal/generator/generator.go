package generator

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	topicNameOptionNumber         = protowire.Number(50001)
	serializationTypeOptionNumber = protowire.Number(50002)

	natsBusImportPath = protogen.GoImportPath("github.com/artarts36/nats-bus")
)

var (
	errMissingDocTemplate = errors.New("doc_template parameter is required")
	errMissingDocOutput   = errors.New("doc_out parameter is required")
)

type Config struct {
	DocTemplatePath string
	DocOutputPath   string
}

type eventMessage struct {
	Message           *protogen.Message
	TopicName         string
	SerializationType string
	Description       string
	TopicConstantName string
}

type topicDocumentationRow struct {
	TopicName   string
	Description string
}

type documentationData struct {
	Topics []topicDocumentationRow
}

type eventOptions struct {
	TopicName         string
	SerializationType string
}

func Generate(plugin *protogen.Plugin, cfg Config) error {
	cfg.DocTemplatePath = strings.TrimSpace(cfg.DocTemplatePath)
	cfg.DocOutputPath = strings.TrimSpace(cfg.DocOutputPath)

	if cfg.DocTemplatePath == "" {
		return errMissingDocTemplate
	}

	if cfg.DocOutputPath == "" {
		return errMissingDocOutput
	}

	docTemplate, err := loadDocumentationTemplate(cfg.DocTemplatePath)
	if err != nil {
		return err
	}

	docRows := make([]topicDocumentationRow, 0)
	for _, file := range plugin.Files {
		if !file.Generate {
			continue
		}

		events, collectErr := collectEventMessages(file)
		if collectErr != nil {
			return collectErr
		}

		if len(events) > 0 {
			generateGoFile(plugin, file, events)
		}

		docRows = append(docRows, buildDocumentationRows(events)...)
	}

	sort.Slice(docRows, func(i int, j int) bool {
		return docRows[i].TopicName < docRows[j].TopicName
	})

	return generateDocumentationFile(plugin, cfg.DocOutputPath, docTemplate, docRows)
}

func loadDocumentationTemplate(path string) (*template.Template, error) {
	templateContent, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read doc template %q: %w", path, err)
	}

	parsedTemplate, err := template.New("topics").Parse(string(templateContent))
	if err != nil {
		return nil, fmt.Errorf("parse doc template %q: %w", path, err)
	}

	return parsedTemplate, nil
}

func collectEventMessages(file *protogen.File) ([]eventMessage, error) {
	events := make([]eventMessage, 0, len(file.Messages))
	if err := appendEventMessages(file.Messages, &events); err != nil {
		return nil, fmt.Errorf("collect events for file %q: %w", file.Desc.Path(), err)
	}

	return events, nil
}

func appendEventMessages(messages []*protogen.Message, events *[]eventMessage) error {
	for _, message := range messages {
		if message.Desc.IsMapEntry() {
			continue
		}

		event, err := buildEventMessage(message)
		if err != nil {
			return err
		}

		*events = append(*events, event)
		if err = appendEventMessages(message.Messages, events); err != nil {
			return err
		}
	}

	return nil
}

func buildEventMessage(message *protogen.Message) (eventMessage, error) {
	options, ok := message.Desc.Options().(*descriptorpb.MessageOptions)
	if !ok || options == nil {
		return eventMessage{}, fmt.Errorf("message %q has no message options", message.Desc.FullName())
	}

	parsedOptions, err := parseEventOptions(options.ProtoReflect().GetUnknown())
	if err != nil {
		return eventMessage{}, fmt.Errorf("message %q: %w", message.Desc.FullName(), err)
	}

	if err = validateEventOptions(parsedOptions, message.Desc.FullName()); err != nil {
		return eventMessage{}, err
	}

	return eventMessage{
		Message:           message,
		TopicName:         parsedOptions.TopicName,
		SerializationType: parsedOptions.SerializationType,
		Description:       normalizeProtoComment(message.Comments.Leading.String()),
		TopicConstantName: "Topic" + message.GoIdent.GoName,
	}, nil
}

func validateEventOptions(options eventOptions, messageName protoreflect.FullName) error {
	if strings.TrimSpace(options.TopicName) == "" {
		return fmt.Errorf("message %q: required option nats_bus.events.topic_name is missing", messageName)
	}

	if strings.TrimSpace(options.SerializationType) == "" {
		return fmt.Errorf("message %q: required option nats_bus.events.serialization_type is missing", messageName)
	}

	return nil
}

func parseEventOptions(rawOptions []byte) (eventOptions, error) {
	parsed := eventOptions{}

	for len(rawOptions) > 0 {
		fieldNumber, wireType, consumedTag := protowire.ConsumeTag(rawOptions)
		if consumedTag < 0 {
			return eventOptions{}, fmt.Errorf("decode message options tag: %w", protowire.ParseError(consumedTag))
		}
		rawOptions = rawOptions[consumedTag:]

		switch fieldNumber {
		case topicNameOptionNumber, serializationTypeOptionNumber:
			if wireType != protowire.BytesType {
				return eventOptions{}, fmt.Errorf(
					"option field %d has invalid wire type %v, expected string",
					fieldNumber,
					wireType,
				)
			}

			value, consumedValue, err := consumeStringValue(rawOptions)
			if err != nil {
				return eventOptions{}, fmt.Errorf("decode option field %d: %w", fieldNumber, err)
			}
			rawOptions = rawOptions[consumedValue:]

			if fieldNumber == topicNameOptionNumber {
				parsed.TopicName = value
			} else {
				parsed.SerializationType = value
			}
		default:
			consumedValue := protowire.ConsumeFieldValue(fieldNumber, wireType, rawOptions)
			if consumedValue < 0 {
				return eventOptions{}, fmt.Errorf(
					"skip message option field %d: %w",
					fieldNumber,
					protowire.ParseError(consumedValue),
				)
			}
			rawOptions = rawOptions[consumedValue:]
		}
	}

	return parsed, nil
}

func consumeStringValue(raw []byte) (string, int, error) {
	value, consumed := protowire.ConsumeBytes(raw)
	if consumed < 0 {
		return "", 0, protowire.ParseError(consumed)
	}

	return string(value), consumed, nil
}

func generateGoFile(plugin *protogen.Plugin, file *protogen.File, events []eventMessage) {
	generatedFile := plugin.NewGeneratedFile(file.GeneratedFilenamePrefix+".event-schemas.pb.go", file.GoImportPath)
	generatedFile.P("// Code generated by protoc-gen-event-schemas. DO NOT EDIT.")
	generatedFile.P("// source: ", file.Desc.Path())
	generatedFile.P()
	generatedFile.P("package ", file.GoPackageName)
	generatedFile.P()

	generateTopicConstants(generatedFile, events)

	for _, event := range events {
		generateMessageStruct(generatedFile, event)
		generateTopicMetaMethod(generatedFile, event)
	}
}

func generateTopicConstants(generatedFile *protogen.GeneratedFile, events []eventMessage) {
	generatedFile.P("// Topic names for generated events.")
	generatedFile.P("const (")
	for _, event := range events {
		generatedFile.P(event.TopicConstantName, " = ", strconv.Quote(event.TopicName))
	}
	generatedFile.P(")")
	generatedFile.P()
}

func generateMessageStruct(generatedFile *protogen.GeneratedFile, event eventMessage) {
	comment := sanitizeGoComment(event.Description)
	if comment != "" {
		generatedFile.P("// ", comment)
	}

	generatedFile.P("type ", event.Message.GoIdent.GoName, " struct {")
	for _, field := range event.Message.Fields {
		fieldComment := sanitizeGoComment(normalizeProtoComment(field.Comments.Leading.String()))
		if fieldComment != "" {
			generatedFile.P("// ", fieldComment)
		}

		generatedFile.P(
			field.GoName,
			" ",
			goTypeForField(generatedFile, field),
			" `json:\"",
			string(field.Desc.Name()),
			",omitempty\"`",
		)
	}
	generatedFile.P("}")
	generatedFile.P()
}

func goTypeForField(generatedFile *protogen.GeneratedFile, field *protogen.Field) string {
	if field.Desc.IsMap() {
		return mapTypeForField(generatedFile, field)
	}

	valueType := scalarGoType(generatedFile, field, true)
	if field.Desc.IsList() {
		return "[]" + valueType
	}

	return valueType
}

func mapTypeForField(generatedFile *protogen.GeneratedFile, field *protogen.Field) string {
	if field.Message == nil || len(field.Message.Fields) < 2 {
		return "map[string]any"
	}

	keyType := scalarGoType(generatedFile, field.Message.Fields[0], false)
	valueType := scalarGoType(generatedFile, field.Message.Fields[1], true)

	return "map[" + keyType + "]" + valueType
}

func scalarGoType(generatedFile *protogen.GeneratedFile, field *protogen.Field, allowMessagePointer bool) string {
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		return "bool"
	case protoreflect.EnumKind:
		if field.Enum == nil {
			return "int32"
		}

		return generatedFile.QualifiedGoIdent(field.Enum.GoIdent)
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int32"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	case protoreflect.FloatKind:
		return "float32"
	case protoreflect.DoubleKind:
		return "float64"
	case protoreflect.StringKind:
		return "string"
	case protoreflect.BytesKind:
		return "[]byte"
	case protoreflect.MessageKind, protoreflect.GroupKind:
		if field.Message == nil {
			return "any"
		}

		messageType := generatedFile.QualifiedGoIdent(field.Message.GoIdent)
		if allowMessagePointer {
			return "*" + messageType
		}

		return messageType
	default:
		return "any"
	}
}

func generateTopicMetaMethod(generatedFile *protogen.GeneratedFile, event eventMessage) {
	topicMetaType := generatedFile.QualifiedGoIdent(protogen.GoIdent{
		GoImportPath: natsBusImportPath,
		GoName:       "TopicMeta",
	})

	generatedFile.P("func (x *", event.Message.GoIdent.GoName, ") TopicMeta() ", topicMetaType, " {")
	generatedFile.P("return ", topicMetaType, "{")
	generatedFile.P("TopicName: ", event.TopicConstantName, ",")
	generatedFile.P("SerializationType: ", serializationTypeExpression(generatedFile, event.SerializationType), ",")
	generatedFile.P("}")
	generatedFile.P("}")
	generatedFile.P()
}

func serializationTypeExpression(generatedFile *protogen.GeneratedFile, value string) string {
	if strings.EqualFold(value, "json") {
		return generatedFile.QualifiedGoIdent(protogen.GoIdent{
			GoImportPath: natsBusImportPath,
			GoName:       "SerializationJSON",
		})
	}

	return generatedFile.QualifiedGoIdent(protogen.GoIdent{
		GoImportPath: natsBusImportPath,
		GoName:       "SerializationType",
	}) + "(" + strconv.Quote(value) + ")"
}

func sanitizeGoComment(text string) string {
	if text == "" {
		return ""
	}

	return strings.Join(strings.Fields(text), " ")
}

func normalizeProtoComment(text string) string {
	lines := strings.Split(text, "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		trimmed = strings.TrimPrefix(trimmed, "//")
		trimmed = strings.TrimSpace(trimmed)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}

	return strings.Join(cleaned, " ")
}

func buildDocumentationRows(events []eventMessage) []topicDocumentationRow {
	rows := make([]topicDocumentationRow, 0, len(events))
	for _, event := range events {
		rows = append(rows, topicDocumentationRow{
			TopicName:   event.TopicName,
			Description: sanitizeMarkdownCell(event.Description),
		})
	}

	return rows
}

func sanitizeMarkdownCell(text string) string {
	normalized := strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
	normalized = strings.ReplaceAll(normalized, "|", "\\|")

	return normalized
}

func generateDocumentationFile(
	plugin *protogen.Plugin,
	outputPath string,
	docTemplate *template.Template,
	rows []topicDocumentationRow,
) error {
	contentBuffer := bytes.Buffer{}
	if err := docTemplate.Execute(&contentBuffer, documentationData{Topics: rows}); err != nil {
		return fmt.Errorf("render documentation template: %w", err)
	}

	docFile := plugin.NewGeneratedFile(outputPath, "")
	writeMultiline(docFile, contentBuffer.String())

	return nil
}

func writeMultiline(generatedFile *protogen.GeneratedFile, content string) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		generatedFile.P(line)
	}
}
