package compile

// pythonKeywords contains all Python 3.8+ reserved keywords
var pythonKeywords = map[string]bool{
	// Boolean values
	"False": true,
	"None":  true,
	"True":  true,

	// Control flow
	"break":    true,
	"continue": true,
	"elif":     true,
	"else":     true,
	"for":      true,
	"if":       true,
	"pass":     true,
	"while":    true,

	// Functions and classes
	"class":  true,
	"def":    true,
	"lambda": true,
	"return": true,
	"yield":  true,

	// Exception handling
	"assert":  true,
	"except":  true,
	"finally": true,
	"raise":   true,
	"try":     true,

	// Imports
	"as":     true,
	"from":   true,
	"import": true,

	// Context managers
	"with": true,

	// Logical operators
	"and": true,
	"in":  true,
	"is":  true,
	"not": true,
	"or":  true,

	// Other
	"async":    true,
	"await":    true,
	"del":      true,
	"global":   true,
	"nonlocal": true,
}

// pythonBuiltins contains common built-in names that should be avoided
var pythonBuiltins = map[string]bool{
	// Types
	"bool":  true,
	"bytes": true,
	"dict":  true,
	"float": true,
	"int":   true,
	"list":  true,
	"set":   true,
	"str":   true,
	"tuple": true,

	// Common functions
	"filter": true,
	"format": true,
	"id":     true,
	"input":  true,
	"len":    true,
	"map":    true,
	"open":   true,
	"print":  true,
	"range":  true,
	"type":   true,
	"zip":    true,
}

// SanitizePythonIdentifier ensures a name is safe to use as a Python identifier
func SanitizePythonIdentifier(name string) string {
	// Check if it's a keyword
	if pythonKeywords[name] {
		return name + "_"
	}

	// Optionally check built-ins (can be configured)
	if pythonBuiltins[name] {
		return name + "_"
	}

	// Check if it starts with a number (Python identifiers can't)
	if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
		return "_" + name
	}

	return name
}

// IsPythonKeyword checks if a string is a Python keyword
func IsPythonKeyword(name string) bool {
	return pythonKeywords[name]
}

// IsPythonBuiltin checks if a string is a Python builtin
func IsPythonBuiltin(name string) bool {
	return pythonBuiltins[name]
}
