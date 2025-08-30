package formatdef

import (
	"fmt"
	"strings"
)

// ContentBuilder helps generate formatted code with proper indentation
type ContentBuilder struct {
	lines       []string
	indentLevel int
	indentStr   string
}

// NewContentBuilder creates a new content builder
func NewContentBuilder(indentStr string) *ContentBuilder {
	if indentStr == "" {
		indentStr = "    " // Default to 4 spaces
	}
	return &ContentBuilder{
		lines:     []string{},
		indentStr: indentStr,
	}
}

// Line adds a line with current indentation
func (b *ContentBuilder) Line(format string, args ...interface{}) *ContentBuilder {
	line := fmt.Sprintf(format, args...)
	if line == "" {
		b.lines = append(b.lines, "")
	} else {
		indent := strings.Repeat(b.indentStr, b.indentLevel)
		b.lines = append(b.lines, indent+line)
	}
	return b
}

// Indent increases indentation level
func (b *ContentBuilder) Indent() *ContentBuilder {
	b.indentLevel++
	return b
}

// Dedent decreases indentation level
func (b *ContentBuilder) Dedent() *ContentBuilder {
	if b.indentLevel > 0 {
		b.indentLevel--
	}
	return b
}

// Comment adds a comment line
func (b *ContentBuilder) Comment(format string, args ...interface{}) *ContentBuilder {
	// Python comment syntax
	return b.Line("# "+format, args...)
}

// BlockComment adds a multi-line comment
func (b *ContentBuilder) BlockComment(lines ...string) *ContentBuilder {
	// Python docstring style
	b.Line(`"""`)
	for _, line := range lines {
		b.Line(line)
	}
	b.Line(`"""`)
	return b
}

// AppendToLastLine appends text to the last line (useful for adding commas, etc)
func (b *ContentBuilder) AppendToLastLine(text string) *ContentBuilder {
	if len(b.lines) > 0 {
		b.lines[len(b.lines)-1] += text
	}
	return b
}

// Build returns the final content as a byte array
func (b *ContentBuilder) Build() []byte {
	return []byte(strings.Join(b.lines, "\n"))
}

// String returns the content as a string
func (b *ContentBuilder) String() string {
	return strings.Join(b.lines, "\n")
}

// Common formatting helpers

// FormatList formats a list of items with a separator
func FormatList(items []string, separator string) string {
	return strings.Join(items, separator)
}

// QuoteString adds quotes around a string (adjust for your format)
func QuoteString(s string) string {
	// TODO: Adjust quote style for your format (single vs double quotes)
	return fmt.Sprintf(`"%s"`, s)
}

// ToPascalCase converts a string to PascalCase
func ToPascalCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, "")
}

// ToCamelCase converts a string to camelCase
func ToCamelCase(s string) string {
	pascal := ToPascalCase(s)
	if len(pascal) > 0 {
		return strings.ToLower(pascal[:1]) + pascal[1:]
	}
	return pascal
}

// ToSnakeCase converts a string to snake_case
func ToSnakeCase(s string) string {
	// Handle special cases
	switch s {
	case "ID":
		return "id"
	case "TaxID":
		return "tax_id"
	case "ContactInfo":
		return "contact_info"
	}

	var result []rune
	for i, r := range s {
		if i > 0 && 'A' <= r && r <= 'Z' {
			// Look at previous character
			prevIsLower := i > 0 && 'a' <= rune(s[i-1]) && rune(s[i-1]) <= 'z'
			// Look at next character if exists
			nextIsLower := i+1 < len(s) && 'a' <= rune(s[i+1]) && rune(s[i+1]) <= 'z'

			// Add underscore before uppercase letter if:
			// 1. Previous char is lowercase (camelCase -> camel_case)
			// 2. This is start of new word in acronym (XMLParser -> xml_parser)
			if prevIsLower || (i > 0 && nextIsLower) {
				result = append(result, '_')
			}
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}
