package formatdef

// Struct represents a structure/class/interface in the target format
// TODO: Adapt this structure to match your target format's type representation
type Struct struct {
	Name   string
	Fields []Field
	// TODO: Add format-specific properties
	// Examples:
	// - Extends string (base class/interface)
	// - Implements []string (interfaces)
	// - IsAbstract bool
	// - Decorators []string
	// - AccessModifier string (public/private/protected)
}

// Field represents a field in a structure
type Field struct {
	Name       string
	Type       Type
	IsOptional bool // When true, generates Optional[T] = None in Python
}

// GetDefinition returns the full struct definition in the target format
func (s *Struct) GetDefinition() string {
	// TODO: Implement format-specific struct/class/interface syntax generation
	// Example for TypeScript: "export interface MyStruct { ... }"
	// Example for Python: "class MyStruct: ..."
	// Example for Java: "public class MyStruct { ... }"
	return "// TODO: Generate " + s.Name + " struct definition"
}
