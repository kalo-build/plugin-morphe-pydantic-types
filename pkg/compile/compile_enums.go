package compile

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kalo-build/morphe-go/pkg/registry"
	"github.com/kalo-build/morphe-go/pkg/yaml"
	"github.com/kalo-build/plugin-morphe-pydantic-types/pkg/formatdef"
)

// CompileEnum converts a Morphe enum to the target format
func CompileEnum(enum yaml.Enum) (*formatdef.Enum, error) {
	// Create the enum definition
	formatEnum := &formatdef.Enum{
		Name:    enum.Name,
		Type:    mapEnumType(enum.Type),
		Entries: make([]formatdef.EnumEntry, 0, len(enum.Entries)),
	}

	// Sort entries for consistent output
	var entryNames []string
	for name := range enum.Entries {
		entryNames = append(entryNames, name)
	}
	sort.Strings(entryNames)

	// Convert each enum entry
	for _, entryName := range entryNames {
		entry := formatdef.EnumEntry{
			Name:  entryName,
			Value: enum.Entries[entryName],
		}
		formatEnum.Entries = append(formatEnum.Entries, entry)
	}

	return formatEnum, nil
}

// mapEnumType maps Morphe enum types to format-specific types
func mapEnumType(morpheType yaml.EnumType) formatdef.Type {
	switch morpheType {
	case yaml.EnumTypeInteger:
		return formatdef.TypeInteger
	case yaml.EnumTypeFloat:
		return formatdef.TypeFloat
	case yaml.EnumTypeString:
		return formatdef.TypeString
	default:
		return formatdef.TypeString
	}
}

// CompileAllEnums compiles all enums and writes them using the writer
func CompileAllEnums(config MorpheCompileConfig, r *registry.Registry, writer *MorpheWriter) error {
	enumContents := make(map[string][]byte)

	// Process each enum in the registry
	for enumName, enum := range r.GetAllEnums() {
		// Compile the enum
		compiledEnum, err := CompileEnum(enum)
		if err != nil {
			return fmt.Errorf("failed to compile enum %s: %w", enumName, err)
		}

		// Generate the content for this enum
		content := generateEnumContent(compiledEnum, config.FormatConfig)
		enumContents[enumName] = content
	}

	// Write all enum contents
	return writer.WriteAllEnums(enumContents)
}

// generateEnumContent generates Python enum definition
func generateEnumContent(enum *formatdef.Enum, config PydanticConfig) []byte {
	cb := formatdef.NewContentBuilder("    ") // 4 spaces for Python

	// Add imports
	cb.Line("from enum import Enum")
	cb.Line("")
	cb.Line("")

	// Generate enum class
	cb.Line("class %s(Enum):", enum.Name)
	cb.Indent()

	// Add docstring
	cb.Line(`"""%s enumeration."""`, enum.Name)

	// Add enum entries
	for _, entry := range enum.Entries {
		// Python enum format: NAME = value
		entryName := formatdef.ToSnakeCase(entry.Name)
		entryName = strings.ToUpper(entryName)

		switch enum.Type.GetName() {
		case "str":
			cb.Line("%s = %q", entryName, entry.Value)
		default:
			cb.Line("%s = %v", entryName, entry.Value)
		}
	}

	// Add utility methods
	cb.Line("")
	cb.Line("@classmethod")
	cb.Line("def from_value(cls, value):")
	cb.Indent()
	cb.Line(`"""Get enum member from value."""`)
	cb.Line("for member in cls:")
	cb.Indent()
	cb.Line("if member.value == value:")
	cb.Indent()
	cb.Line("return member")
	cb.Dedent()
	cb.Dedent()
	cb.Line("raise ValueError(f\"No %s member with value {value}\")", enum.Name)
	cb.Dedent()

	return cb.Build()
}
