package compile

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kalo-build/morphe-go/pkg/registry"
	"github.com/kalo-build/morphe-go/pkg/yaml"
	"github.com/kalo-build/morphe-go/pkg/yamlops"
	"github.com/kalo-build/plugin-morphe-pydantic-types/pkg/compile/cfg"
	"github.com/kalo-build/plugin-morphe-pydantic-types/pkg/formatdef"
	"github.com/kalo-build/plugin-morphe-pydantic-types/pkg/typemap"
)

// resolvePolymorphicThrough looks up the model that has the polymorphic relationship
func resolvePolymorphicThrough(through string, r *registry.Registry) (string, error) {
	// Find the model that has this polymorphic relationship
	for modelName, model := range r.GetAllModels() {
		for relName, rel := range model.Related {
			if relName == through && yamlops.IsRelationPoly(string(rel.Type)) {
				return modelName, nil
			}
		}
	}
	return "", fmt.Errorf("polymorphic relationship %s not found", through)
}

// CompileModel converts a Morphe model to the target format
func CompileModel(model yaml.Model, r *registry.Registry) (*formatdef.Struct, error) {
	// Create the struct definition
	formatStruct := &formatdef.Struct{
		Name:   model.Name,
		Fields: make([]formatdef.Field, 0),
	}

	// Sort fields for consistent output
	var fieldNames []string
	for name := range model.Fields {
		fieldNames = append(fieldNames, name)
	}
	sort.Strings(fieldNames)

	// Add fields
	for _, fieldName := range fieldNames {
		field := model.Fields[fieldName]
		fieldType := typemap.GetFieldType(field.Type)
		formatField := formatdef.Field{
			Name: fieldName,
			Type: fieldType,
		}
		formatStruct.Fields = append(formatStruct.Fields, formatField)
	}

	// Process related models (if any)
	if len(model.Related) > 0 {
		// Sort related for consistent output
		var relatedNames []string
		for name := range model.Related {
			relatedNames = append(relatedNames, name)
		}
		sort.Strings(relatedNames)

		// Add foreign key fields
		for _, relatedName := range relatedNames {
			relation := model.Related[relatedName]
			relationType := string(relation.Type)

			// Handle polymorphic relationships
			if yamlops.IsRelationPoly(relationType) && yamlops.IsRelationFor(relationType) && yamlops.IsRelationOne(relationType) {
				// ForOnePoly: Add type and id fields
				typeField := formatdef.Field{
					Name: formatdef.ToCamelCase(relatedName + "_type"),
					Type: formatdef.TypeString,
				}
				formatStruct.Fields = append(formatStruct.Fields, typeField)

				idField := formatdef.Field{
					Name: formatdef.ToCamelCase(relatedName + "_id"),
					Type: formatdef.TypeString,
				}
				formatStruct.Fields = append(formatStruct.Fields, idField)
			} else if yamlops.IsRelationPoly(relationType) {
				// Other polymorphic types (HasOnePoly, HasManyPoly, ForManyPoly)
				// These don't add fields to the model, but affect how we handle relationships
				continue
			} else if yamlops.IsRelationFor(relationType) && yamlops.IsRelationOne(relationType) {
				// Regular ForOne: Add foreign key field
				relField := formatdef.Field{
					Name: formatdef.ToCamelCase(relatedName + "_id"),
					Type: formatdef.TypeString,
				}
				formatStruct.Fields = append(formatStruct.Fields, relField)
			}
			// HasOne, HasMany, ForMany don't add fields to this model
		}

		// Add navigation properties for relationships (for Python type hints)
		for _, relatedName := range relatedNames {
			relation := model.Related[relatedName]
			relationType := string(relation.Type)

			// Resolve the actual target model name using aliasing
			targetModelName := yamlops.GetRelationTargetName(relatedName, relation.Aliased)

			// Add navigation field based on relationship type
			var navType formatdef.Type
			if yamlops.IsRelationPoly(relationType) {
				// Polymorphic relationships need Union types
				if len(relation.For) > 0 {
					// Create a custom type representing the Union
					unionType := "Union["
					for i, forModel := range relation.For {
						if i > 0 {
							unionType += ", "
						}
						unionType += "'" + forModel + "'"
					}
					unionType += "]"
					navType = formatdef.BasicType{Name: unionType}
				} else if relation.Through != "" {
					// HasManyPoly/HasOnePoly with through - resolve the actual model
					throughModel, err := resolvePolymorphicThrough(relation.Through, r)
					if err != nil {
						// Fallback to Any if we can't resolve
						navType = formatdef.TypeAny
					} else {
						navType = formatdef.BasicType{Name: throughModel}
					}
				} else {
					// No 'for' or 'through' specified, use Any
					navType = formatdef.TypeAny
				}
			} else {
				// Regular relationship
				navType = formatdef.BasicType{Name: targetModelName}
			}

			// Determine if it's a collection
			if yamlops.IsRelationMany(relationType) {
				navType = formatdef.ArrayType{ElementType: navType}
			}

			// Add navigation field (prefixed with _ to distinguish from data fields)
			navField := formatdef.Field{
				Name: "_nav_" + relatedName,
				Type: navType,
			}
			formatStruct.Fields = append(formatStruct.Fields, navField)
		}
	}

	return formatStruct, nil
}

// CompileAllModels compiles all models and writes them using the writer
func CompileAllModels(config MorpheCompileConfig, r *registry.Registry, writer *MorpheWriter) error {
	modelContents := make(map[string][]byte)

	// Process each model in the registry
	for modelName, model := range r.GetAllModels() {
		// Compile the model
		compiledModel, err := CompileModel(model, r)
		if err != nil {
			return fmt.Errorf("failed to compile model %s: %w", modelName, err)
		}

		// Generate the content for this model
		content := generateModelContent(compiledModel, config.FormatConfig, config.MorpheConfig, r)
		modelContents[modelName] = content
	}

	// Write all model contents
	return writer.WriteAllModels(modelContents)
}

// generateModelContent generates Python Pydantic model
func generateModelContent(model *formatdef.Struct, config PydanticConfig, morpheConfig cfg.MorpheConfig, r *registry.Registry) []byte {
	cb := formatdef.NewContentBuilder("    ")

	// Create import tracker
	imports := NewImportTracker(r)

	// Add Pydantic imports
	imports.AddPydantic("BaseModel")
	if morpheConfig.Models.UseField {
		imports.AddPydantic("Field")
	}

	// Track whether we need model config
	needsModelConfig := false
	hasPolymorphicTypeField := false
	polymorphicTypeToNavMap := make(map[string]string)

	// Scan all fields to determine imports
	for _, field := range model.Fields {
		// Skip navigation properties
		if strings.HasPrefix(field.Name, "_nav_") {
			continue
		}

		typeName := field.Type.GetName()
		imports.TrackFieldType(typeName)

		// Check if this field is an enum
		if basicType, ok := field.Type.(formatdef.BasicType); ok {
			innerType := extractInnerType(basicType.Name)
			if innerType != "" && resolveFieldType(innerType, r) == "enum" {
				needsModelConfig = true
			}
		}

		// Check for polymorphic type fields
		if strings.HasSuffix(field.Name, "_type") && typeName == "str" {
			// Look for corresponding nav field
			navFieldName := "_nav_" + strings.TrimSuffix(field.Name, "_type")
			polymorphicTypeToNavMap[field.Name] = navFieldName
			hasPolymorphicTypeField = true
		}
	}

	// Scan navigation properties
	for _, field := range model.Fields {
		if !strings.HasPrefix(field.Name, "_nav_") {
			continue
		}

		typeName := field.Type.GetName()
		imports.TrackFieldType(typeName)
	}

	// We always need Optional for navigation properties
	if config.AddTypeHints {
		imports.AddTyping("Optional")
	}

	// Add Literal if we have polymorphic type fields
	if hasPolymorphicTypeField {
		imports.AddTyping("Literal")
	}

	// Generate imports
	imports.Generate(cb)
	cb.Line("")

	// Generate class
	cb.Line("class %s(BaseModel):", model.Name)
	cb.Indent()

	// Add docstring
	cb.Line(`"""%s model."""`, model.Name)

	if len(model.Fields) == 0 {
		cb.Line("pass")
	} else {
		// Add fields
		for _, field := range model.Fields {
			// Skip navigation properties
			if strings.HasPrefix(field.Name, "_nav_") {
				continue
			}

			fieldName := SanitizePythonIdentifier(formatdef.ToSnakeCase(field.Name))
			fieldType := field.Type.GetName()

			// Add type hint
			if config.AddTypeHints {
				// Check if this is a polymorphic type field
				if navFieldName, isPolyType := polymorphicTypeToNavMap[field.Name]; isPolyType {
					// Look for the navigation field to get allowed types
					var allowedTypes []string
					for _, navField := range model.Fields {
						if navField.Name == navFieldName {
							// Extract the types from Union[...]
							unionType := navField.Type.GetName()
							if strings.HasPrefix(unionType, "Union[") && strings.HasSuffix(unionType, "]") {
								unionContent := unionType[6 : len(unionType)-1]
								types := strings.Split(unionContent, ", ")
								for _, t := range types {
									// Remove quotes
									t = strings.Trim(t, "'\"")
									allowedTypes = append(allowedTypes, fmt.Sprintf("\"%s\"", t))
								}
							}
							break
						}
					}
					if len(allowedTypes) > 0 {
						cb.Line("%s: Literal[%s]", fieldName, strings.Join(allowedTypes, ", "))
					} else {
						cb.Line("%s: str", fieldName)
					}
				} else if len(fieldName) > 3 && (fieldName[len(fieldName)-3:] == "_id" || strings.HasSuffix(fieldName, "_type")) {
					// Make foreign keys and type fields optional by default
					cb.Line("%s: Optional[%s] = None", fieldName, fieldType)
				} else {
					cb.Line("%s: %s", fieldName, fieldType)
				}
			} else {
				cb.Line("%s = None", fieldName)
			}
		}

		// Add navigation properties (relationships)
		for _, field := range model.Fields {
			if !strings.HasPrefix(field.Name, "_nav_") {
				continue
			}

			// Remove _nav_ prefix to get the actual relationship name
			relName := strings.TrimPrefix(field.Name, "_nav_")
			fieldName := SanitizePythonIdentifier(formatdef.ToSnakeCase(relName))
			fieldType := field.Type.GetName()

			// Skip if this is a polymorphic relationship with corresponding type/id fields
			hasPolyFields := false
			for _, f := range model.Fields {
				if f.Name == relName+"_type" || f.Name == relName+"_id" {
					hasPolyFields = true
					break
				}
			}

			if hasPolyFields {
				// For polymorphic relationships, add a property that returns the actual object
				// This would typically be implemented with a validator or custom getter
				continue
			}

			// For regular relationships, add the navigation property
			if strings.HasPrefix(fieldType, "List[") {
				// Many relationship - optional list with default empty list
				cb.Line("%s: Optional[%s] = None", fieldName, fieldType)
			} else if strings.Contains(fieldType, "Union[") {
				// Union type - don't add extra quotes
				cb.Line("%s: Optional[%s] = None", fieldName, fieldType)
			} else {
				// One relationship - optional with forward reference
				cb.Line("%s: Optional['%s'] = None", fieldName, fieldType)
			}
		}

		if config.PydanticV2 && needsModelConfig {
			// Add Pydantic v2 model config only if needed
			cb.Line("")
			cb.Line("model_config = {")
			cb.Indent()
			cb.Line(`"validate_assignment": True,`)
			cb.Line(`"use_enum_values": True,`)
			cb.Dedent()
			cb.Line("}")
		} else if !config.PydanticV2 && needsModelConfig {
			// Add Pydantic v1 Config
			cb.Line("")
			cb.Line("class Config:")
			cb.Indent()
			cb.Line("validate_assignment = True")
			cb.Line("use_enum_values = True")
			cb.Dedent()
		}
	}

	cb.Dedent() // End of class body

	return cb.Build()
}
