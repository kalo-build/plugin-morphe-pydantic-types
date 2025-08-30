package formatdef

// Enum represents an enum definition in the target format
// TODO: Adapt this structure to match your target format's enum representation
type Enum struct {
	Name    string
	Type    Type // The underlying type (string, int, etc.)
	Entries []EnumEntry
	// TODO: Add format-specific enum properties
	// Examples:
	// - IsConstEnum bool (for TypeScript)
	// - BaseClass string (for Python)
	// - Package string (for Java)
}

// EnumEntry represents a single enum value
type EnumEntry struct {
	Name  string
	Value interface{}
	// TODO: Add format-specific entry properties
	// Examples:
	// - Comment string (for documentation)
	// - Deprecated bool
	// - Metadata map[string]interface{}
}

// GetDefinition returns the full enum definition in the target format
func (e *Enum) GetDefinition() string {
	// TODO: Implement format-specific enum syntax generation
	// Example for TypeScript: "export enum MyEnum { ... }"
	// Example for Python: "class MyEnum(Enum): ..."
	// Example for Java: "public enum MyEnum { ... }"
	return "// TODO: Generate " + e.Name + " enum definition"
}
