package formatdef

// Type represents a type in the target format
// TODO: Replace with your target format's type system
type Type interface {
	// GetName returns the type name in the target format
	GetName() string
	// IsNullable returns whether the type can be null/nil/undefined
	IsNullable() bool
	// TODO: Add format-specific methods as needed
}

// BasicType represents a basic/primitive type
type BasicType struct {
	Name     string
	Nullable bool
}

func (t BasicType) GetName() string {
	return t.Name
}

func (t BasicType) IsNullable() bool {
	return t.Nullable
}

// ArrayType represents an array/list type
type ArrayType struct {
	ElementType Type
	// TODO: Add format-specific array properties
}

func (t ArrayType) GetName() string {
	// Python list syntax
	return "List[" + t.ElementType.GetName() + "]"
}

func (t ArrayType) IsNullable() bool {
	return false
}

// Python basic types
var (
	TypeString  = BasicType{Name: "str"}
	TypeInteger = BasicType{Name: "int"}
	TypeFloat   = BasicType{Name: "float"}
	TypeBoolean = BasicType{Name: "bool"}
	TypeDate    = BasicType{Name: "datetime"}
	TypeJSON    = BasicType{Name: "Dict[str, Any]"}
	TypeAny     = BasicType{Name: "Any"}
)
