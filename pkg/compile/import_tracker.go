package compile

import (
	"sort"
	"strings"

	"github.com/kalo-build/morphe-go/pkg/registry"
	"github.com/kalo-build/plugin-morphe-pydantic-types/pkg/formatdef"
)

// ImportTracker tracks required imports for Python code generation
type ImportTracker struct {
	pydantic []string
	typing   []string
	datetime bool
	enums    map[string]bool
	models   map[string]bool
	registry *registry.Registry
}

// NewImportTracker creates a new import tracker
func NewImportTracker(r *registry.Registry) *ImportTracker {
	return &ImportTracker{
		enums:    make(map[string]bool),
		models:   make(map[string]bool),
		registry: r,
	}
}

// AddPydantic adds a pydantic import
func (it *ImportTracker) AddPydantic(imports ...string) {
	for _, imp := range imports {
		if !containsString(it.pydantic, imp) {
			it.pydantic = append(it.pydantic, imp)
		}
	}
}

// TrackFieldType analyzes a field type and tracks necessary imports
func (it *ImportTracker) TrackFieldType(typeName string) {
	// Check for typing imports
	if strings.Contains(typeName, "Optional[") {
		it.AddTyping("Optional")
	}
	if strings.Contains(typeName, "List[") {
		it.AddTyping("List")
	}
	if strings.Contains(typeName, "Union[") {
		it.AddTyping("Union")
	}
	if strings.Contains(typeName, "Dict[") {
		it.AddTyping("Dict")
	}
	if strings.Contains(typeName, "Any") {
		it.AddTyping("Any")
	}
	if strings.Contains(typeName, "Literal[") {
		it.AddTyping("Literal")
	}

	// Check for datetime
	if typeName == "datetime" || strings.Contains(typeName, "datetime") {
		it.datetime = true
	}

	// Extract inner types and check if they're enums or models
	innerTypes := extractAllInnerTypes(typeName)
	for _, innerType := range innerTypes {
		if innerType != "" && !isBasicType(innerType) {
			switch resolveFieldType(innerType, it.registry) {
			case "enum":
				it.enums[innerType] = true
			case "model":
				it.models[innerType] = true
				it.AddTyping("TYPE_CHECKING")
			}
		}
	}
}

// AddTyping adds a typing import
func (it *ImportTracker) AddTyping(imports ...string) {
	for _, imp := range imports {
		if !containsString(it.typing, imp) {
			it.typing = append(it.typing, imp)
		}
	}
}

// Generate generates the import statements
func (it *ImportTracker) Generate(cb *formatdef.ContentBuilder) {
	// Pydantic imports
	if len(it.pydantic) > 0 {
		cb.Line("from pydantic import %s", strings.Join(it.pydantic, ", "))
	}

	// Typing imports
	if len(it.typing) > 0 {
		sort.Strings(it.typing)
		cb.Line("from typing import %s", strings.Join(it.typing, ", "))
	}

	// Datetime
	if it.datetime {
		cb.Line("from datetime import datetime")
	}

	// Enums
	if len(it.enums) > 0 {
		var enumNames []string
		for enum := range it.enums {
			enumNames = append(enumNames, enum)
		}
		sort.Strings(enumNames)
		for _, enumName := range enumNames {
			cb.Line("from ..enums.%s import %s", formatdef.ToSnakeCase(enumName), enumName)
		}
	}

	cb.Line("")

	// Models under TYPE_CHECKING
	if len(it.models) > 0 {
		cb.Line("if TYPE_CHECKING:")
		cb.Indent()
		var modelNames []string
		for model := range it.models {
			modelNames = append(modelNames, model)
		}
		sort.Strings(modelNames)
		for _, modelName := range modelNames {
			cb.Line("from .%s import %s", formatdef.ToSnakeCase(modelName), modelName)
		}
		cb.Dedent()
	}
}

// Helper functions

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func isBasicType(typeName string) bool {
	basicTypes := []string{"str", "int", "float", "bool", "datetime", "Any", "None"}
	for _, basic := range basicTypes {
		if typeName == basic {
			return true
		}
	}
	return false
}

// extractAllInnerTypes extracts all type names from complex type expressions
func extractAllInnerTypes(typeName string) []string {
	var types []string

	// Remove all brackets and split by comma
	cleaned := typeName
	cleaned = strings.ReplaceAll(cleaned, "[", " ")
	cleaned = strings.ReplaceAll(cleaned, "]", " ")
	cleaned = strings.ReplaceAll(cleaned, ",", " ")

	// Split and clean each part
	parts := strings.Fields(cleaned)
	for _, part := range parts {
		part = strings.Trim(part, "'\"")
		if part != "" && !strings.Contains(part, "|") {
			types = append(types, part)
		}
	}

	return types
}
