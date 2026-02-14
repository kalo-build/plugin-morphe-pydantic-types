package compile

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kalo-build/morphe-go/pkg/registry"
	"github.com/kalo-build/morphe-go/pkg/yaml"
	"github.com/kalo-build/morphe-go/pkg/yamlops"
	"github.com/kalo-build/plugin-morphe-pydantic-types/pkg/formatdef"
	"github.com/kalo-build/plugin-morphe-pydantic-types/pkg/typemap"
)

// CompileEntity converts a Morphe entity to the target format
func CompileEntity(entity yaml.Entity, r *registry.Registry) (*formatdef.Struct, error) {
	// Create the struct definition
	formatStruct := &formatdef.Struct{
		Name:   entity.Name,
		Fields: make([]formatdef.Field, 0),
	}

	// Sort fields for consistent output
	var fieldNames []string
	for name := range entity.Fields {
		fieldNames = append(fieldNames, name)
	}
	sort.Strings(fieldNames)

	// Process entity fields
	for _, fieldName := range fieldNames {
		field := entity.Fields[fieldName]
		fieldType, err := resolveEntityFieldType(field.Type, r)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve field type for %s: %w", fieldName, err)
		}

		formatField := formatdef.Field{
			Name:       fieldName,
			Type:       fieldType,
			IsOptional: hasAttribute(field.Attributes, "optional"),
		}
		formatStruct.Fields = append(formatStruct.Fields, formatField)
	}

	// Sort and process relationships
	if len(entity.Related) > 0 {
		var relatedNames []string
		for name := range entity.Related {
			relatedNames = append(relatedNames, name)
		}
		sort.Strings(relatedNames)

		for _, relatedName := range relatedNames {
			relation := entity.Related[relatedName]
			relationType := string(relation.Type)

			// Handle polymorphic relationships
			if yamlops.IsRelationPoly(relationType) && yamlops.IsRelationFor(relationType) && yamlops.IsRelationOne(relationType) {
				// ForOnePoly: Add type and id fields
				typeField := formatdef.Field{
					Name: formatdef.ToSnakeCase(relatedName + "_type"),
					Type: formatdef.TypeString,
				}
				formatStruct.Fields = append(formatStruct.Fields, typeField)

				idField := formatdef.Field{
					Name: formatdef.ToSnakeCase(relatedName + "_id"),
					Type: formatdef.TypeString,
				}
				formatStruct.Fields = append(formatStruct.Fields, idField)
			} else if yamlops.IsRelationPoly(relationType) {
				// Other polymorphic types (HasOnePoly, HasManyPoly, ForManyPoly)
				// These don't add fields to the entity
				continue
			} else if yamlops.IsRelationFor(relationType) && yamlops.IsRelationOne(relationType) {
				// Regular ForOne: Add foreign key field
				fkField := formatdef.Field{
					Name: formatdef.ToSnakeCase(relatedName + "_id"),
					Type: formatdef.TypeString,
				}
				formatStruct.Fields = append(formatStruct.Fields, fkField)
			}

			// Add navigation field based on relation type
			// Resolve the actual target using aliasing
			targetName := yamlops.GetRelationTargetName(relatedName, relation.Aliased)

			var navType formatdef.Type
			var navFieldName string

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
				} else {
					// No 'for' specified, use Any
					navType = formatdef.TypeAny
				}
			} else {
				// Regular relationship - use the aliased target
				navType = formatdef.BasicType{Name: targetName}
			}

			// Determine if it's a collection
			if yamlops.IsRelationMany(relationType) {
				if _, ok := navType.(formatdef.BasicType); ok {
					// Wrap in List for many relationships
					navType = formatdef.ArrayType{ElementType: navType}
				}
				// Pluralize for many relationships
				navFieldName = formatdef.ToSnakeCase(relatedName) + "s"
			} else {
				navFieldName = formatdef.ToSnakeCase(relatedName)
			}

			navField := formatdef.Field{
				Name: navFieldName,
				Type: navType,
			}
			formatStruct.Fields = append(formatStruct.Fields, navField)
		}
	}

	return formatStruct, nil
}

// resolveEntityFieldType resolves a model field path to a concrete type
func resolveEntityFieldType(fieldPath yaml.ModelFieldPath, r *registry.Registry) (formatdef.Type, error) {
	// Split the path (e.g., "User.email" or "User.ContactInfo.email")
	parts := strings.Split(string(fieldPath), ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid field path: %s", fieldPath)
	}

	// Get the root model
	currentModel, err := r.GetModel(parts[0])
	if err != nil {
		return nil, fmt.Errorf("model %s not found", parts[0])
	}

	// Navigate through the path
	for i := 1; i < len(parts)-1; i++ {
		// This is a related model
		relation, exists := currentModel.Related[parts[i]]
		if !exists {
			return nil, fmt.Errorf("relation %s not found in model %s", parts[i], currentModel.Name)
		}

		// Resolve the actual target model name using aliasing
		targetModelName := yamlops.GetRelationTargetName(parts[i], relation.Aliased)

		// Get the related model using the resolved target name
		currentModel, err = r.GetModel(targetModelName)
		if err != nil {
			return nil, fmt.Errorf("related model %s not found", targetModelName)
		}
	}

	// Get the terminal field
	fieldName := parts[len(parts)-1]
	field, exists := currentModel.Fields[fieldName]
	if !exists {
		return nil, fmt.Errorf("field %s not found in model %s", fieldName, currentModel.Name)
	}

	// Return the appropriate type
	return typemap.GetFieldType(field.Type), nil
}

// resolveFieldType checks if a type name is an enum, model, or basic type
func resolveFieldType(typeName string, r *registry.Registry) string {
	// Check if it's an enum
	if _, err := r.GetEnum(typeName); err == nil {
		return "enum"
	}
	// Check if it's a model
	if _, err := r.GetModel(typeName); err == nil {
		return "model"
	}
	// Otherwise it's a basic type or unknown
	return "basic"
}

// addToStringSlice adds a string to slice if not already present
func addToStringSlice(slice *[]string, value string) {
	for _, v := range *slice {
		if v == value {
			return
		}
	}
	*slice = append(*slice, value)
}

// extractInnerType extracts the inner type from Optional[X], List[X], etc.
func extractInnerType(typeName string) string {
	// Handle Optional[X], List[X], etc.
	if start := strings.Index(typeName, "["); start != -1 {
		if end := strings.LastIndex(typeName, "]"); end > start {
			inner := typeName[start+1 : end]
			// Remove quotes if present
			inner = strings.Trim(inner, "'\"")
			// If it's a Union, we can't extract a single type
			if strings.Contains(inner, ",") {
				return ""
			}
			return inner
		}
	}
	return typeName
}

// CompileAllEntities compiles all entities and writes them using the writer
func CompileAllEntities(config MorpheCompileConfig, r *registry.Registry, writer *MorpheWriter) error {
	entityContents := make(map[string][]byte)

	// Process each entity in the registry
	for entityName, entity := range r.GetAllEntities() {
		// Compile the entity
		compiledEntity, err := CompileEntity(entity, r)
		if err != nil {
			return fmt.Errorf("failed to compile entity %s: %w", entityName, err)
		}

		// Generate the content for this entity
		content := generateEntityContent(compiledEntity, entity, config.FormatConfig, r)
		entityContents[entityName] = content
	}

	// Write all entity contents
	return writer.WriteAllEntities(entityContents)
}

// generateEntityContent generates Python entity with relationships and identifiers
func generateEntityContent(entity *formatdef.Struct, morpheEntity yaml.Entity, config PydanticConfig, r *registry.Registry) []byte {
	cb := formatdef.NewContentBuilder("    ")

	// Create import tracker
	imports := NewImportTracker(r)

	// Add Pydantic imports
	imports.AddPydantic("BaseModel")
	if config.PydanticV2 {
		imports.AddPydantic("Field")
	}

	// Track if we have polymorphic type fields
	hasPolymorphicTypeField := false

	// Scan all fields to determine imports
	for _, field := range entity.Fields {
		typeName := field.Type.GetName()
		imports.TrackFieldType(typeName)

		// Check for polymorphic type fields
		if strings.HasSuffix(field.Name, "_type") && typeName == "str" {
			relName := strings.TrimSuffix(field.Name, "_type")
			relName = strings.TrimSuffix(relName, "_Type")
			if _, exists := morpheEntity.Related[relName]; exists {
				hasPolymorphicTypeField = true
			}
		}
	}

	// We always need these for entities
	imports.AddTyping("Optional", "List", "TYPE_CHECKING")

	// Add Literal if we have polymorphic type fields
	if hasPolymorphicTypeField {
		imports.AddTyping("Literal")
	}

	// Generate imports
	imports.Generate(cb)
	cb.Line("")

	// Generate class
	cb.Line("class %s(BaseModel):", entity.Name)
	cb.Indent()

	// Add docstring
	cb.BlockComment(
		fmt.Sprintf("%s entity.", entity.Name),
		"",
		fmt.Sprintf("Identifiers: %d", len(morpheEntity.Identifiers)),
		fmt.Sprintf("Relationships: %d", len(morpheEntity.Related)),
	)

	// Group fields by whether they're identifiers
	identifierFields := make(map[string]string)
	for idName, identifier := range morpheEntity.Identifiers {
		for _, fieldName := range identifier.Fields {
			identifierFields[fieldName] = idName
		}
	}

	// Add fields
	for _, field := range entity.Fields {
		fieldName := SanitizePythonIdentifier(formatdef.ToSnakeCase(field.Name))
		fieldType := field.Type.GetName()

		// Add identifier comment
		if idType, isIdentifier := identifierFields[field.Name]; isIdentifier {
			cb.Line("# %s identifier", idType)
		}

		if config.AddTypeHints {
			// Check if this is a polymorphic type field
			if strings.HasSuffix(field.Name, "_type") && fieldType == "str" {
				// Look for the corresponding relationship to get allowed types
				relName := strings.TrimSuffix(field.Name, "_type")
				relName = strings.TrimSuffix(relName, "_Type") // Handle both cases

				if relation, exists := morpheEntity.Related[relName]; exists && len(relation.For) > 0 {
					// Build Literal type with allowed values
					var allowedTypes []string
					for _, forModel := range relation.For {
						allowedTypes = append(allowedTypes, fmt.Sprintf("\"%s\"", forModel))
					}
					cb.Line("%s: Literal[%s]", fieldName, strings.Join(allowedTypes, ", "))
				} else {
					cb.Line("%s: str", fieldName)
				}
			} else if strings.HasPrefix(fieldType, "Optional[") || strings.HasPrefix(fieldType, "List[") || strings.Contains(fieldType, "Union[") {
				// Relationship fields or Union types
				cb.Line("%s: %s = None", fieldName, fieldType)
			} else if field.IsOptional || strings.HasSuffix(fieldName, "_id") || strings.HasSuffix(fieldName, "_type") {
				// Optional attribute, foreign keys, or type fields
				cb.Line("%s: Optional[%s] = None", fieldName, fieldType)
			} else {
				cb.Line("%s: %s", fieldName, fieldType)
			}
		} else {
			cb.Line("%s = None", fieldName)
		}
	}

	// Add identifier methods
	if primary, hasPrimary := morpheEntity.Identifiers["primary"]; hasPrimary && len(primary.Fields) > 0 {
		cb.Line("")
		cb.Line("def get_id(self) -> str:")
		cb.Indent()
		cb.Line(`"""Get the primary identifier."""`)
		cb.Line("return self.%s", formatdef.ToSnakeCase(primary.Fields[0]))
		cb.Dedent()
	}

	// Add relationship loader methods
	if len(morpheEntity.Related) > 0 {
		for relName, relation := range morpheEntity.Related {
			cb.Line("")
			switch relation.Type {
			case "HasMany", "ForMany":
				// Use plural form for method name
				cb.Line("async def load_%ss(self) -> List['%s']:", SanitizePythonIdentifier(formatdef.ToSnakeCase(relName)), relName)
				cb.Indent()
				cb.Line(`"""Load related %s entities."""`, relName)
				cb.Line("# TODO: Implement lazy loading")
				cb.Line("return []")
				cb.Dedent()
			default:
				cb.Line("async def load_%s(self) -> Optional['%s']:", SanitizePythonIdentifier(formatdef.ToSnakeCase(relName)), relName)
				cb.Indent()
				cb.Line(`"""Load related %s entity."""`, relName)
				cb.Line("# TODO: Implement lazy loading")
				cb.Line("return None")
				cb.Dedent()
			}
		}
	}

	if config.PydanticV2 {
		// Add Pydantic v2 model config
		cb.Line("")
		cb.Line("model_config = {")
		cb.Indent()
		cb.Line(`"validate_assignment": True,`)
		cb.Line(`"arbitrary_types_allowed": True,`)
		cb.Dedent()
		cb.Line("}")
	} else {
		// Add Pydantic v1 Config
		cb.Line("")
		cb.Line("class Config:")
		cb.Indent()
		cb.Line("validate_assignment = True")
		cb.Line("arbitrary_types_allowed = True")
		cb.Dedent()
	}

	return cb.Build()
}
