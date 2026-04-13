package main

import (
	"flag"
	"strings"

	"github.com/artarts36/nats-bus/protoc-gen-event-schemas/internal/generator"
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	var flags flag.FlagSet

	docTemplatePath := flags.String("doc_template", "", "path to markdown documentation template")
	docTemplatePathAlt := flags.String("doc-template", "", "path to markdown documentation template")
	docOutputPath := flags.String("doc_out", "", "path to generated markdown documentation file")
	docOutputPathAlt := flags.String("doc-output", "", "path to generated markdown documentation file")

	opts := protogen.Options{ParamFunc: flags.Set}
	opts.Run(func(plugin *protogen.Plugin) error {
		return generator.Generate(plugin, generator.Config{
			DocTemplatePath: firstNonEmpty(*docTemplatePath, *docTemplatePathAlt),
			DocOutputPath:   firstNonEmpty(*docOutputPath, *docOutputPathAlt),
		})
	})
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}

	return ""
}
