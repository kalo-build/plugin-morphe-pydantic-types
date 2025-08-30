package typemap

import (
	"github.com/kalo-build/morphe-go/pkg/registry"
	"github.com/kalo-build/morphe-go/pkg/yaml"
	"github.com/kalo-build/plugin-morphe-pydantic-types/pkg/formatdef"
)

// MorpheModelFieldToFormatType maps Morphe field types to target format types
// TODO: Rename this variable to match your format (e.g., MorpheModelFieldToPythonType)
// TODO: Update the type mappings to match your target format's type system
var MorpheModelFieldToFormatType = map[yaml.ModelFieldType]formatdef.Type{
	// String types
	yaml.ModelFieldTypeString:    formatdef.TypeString,
	yaml.ModelFieldTypeUUID:      formatdef.TypeString,
	yaml.ModelFieldTypeProtected: formatdef.TypeString,
	yaml.ModelFieldTypeSealed:    formatdef.TypeString,

	// Numeric types
	yaml.ModelFieldTypeInteger:       formatdef.TypeInteger,
	yaml.ModelFieldTypeAutoIncrement: formatdef.TypeInteger,
	yaml.ModelFieldTypeFloat:         formatdef.TypeFloat,

	// Boolean type
	yaml.ModelFieldTypeBoolean: formatdef.TypeBoolean,

	// Date/Time types
	yaml.ModelFieldTypeTime: formatdef.TypeDate,
	yaml.ModelFieldTypeDate: formatdef.TypeDate,

	// TODO: Add mappings for any custom field types used in your Morphe schemas
}

// GetFieldType returns the format type for a given Morphe field type
func GetFieldType(fieldType yaml.ModelFieldType) formatdef.Type {
	if formatType, exists := MorpheModelFieldToFormatType[fieldType]; exists {
		return formatType
	}
	// Check if it's an enum type (custom type not in the predefined list)
	// In Morphe, enum references are just the enum name
	// For Python, we'll treat them as the enum type itself
	return formatdef.BasicType{Name: string(fieldType)}
}

// MorpheStructureFieldToFormatType maps structure field types to format types
func MorpheStructureFieldToFormatType(fieldType yaml.StructureFieldType, fieldName string, r *registry.Registry) (formatdef.Type, error) {
	// Structure fields use the same type mappings as model fields
	modelFieldType := yaml.ModelFieldType(fieldType)
	return GetFieldType(modelFieldType), nil
}
