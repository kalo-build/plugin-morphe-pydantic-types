package compile

import (
	"fmt"

	"github.com/kalo-build/morphe-go/pkg/registry"
	"github.com/kalo-build/morphe-go/pkg/yaml"
)

// MorpheToPydantic compiles a Morphe registry to Python with Pydantic models
func MorpheToPydantic(config MorpheCompileConfig) error {
	// Load the Morphe registry
	r, rErr := registry.LoadMorpheRegistry(registry.LoadMorpheRegistryHooks{}, config.MorpheLoadRegistryConfig)
	if rErr != nil {
		return fmt.Errorf("failed to load morphe registry: %w", rErr)
	}

	// Initialize the writer
	writer := NewMorpheWriter(config.OutputPath)

	// Process enums if present
	if r.HasEnums() {
		fmt.Println("Compiling enums...")
		if err := CompileAllEnums(config, r, writer); err != nil {
			return fmt.Errorf("failed to compile enums: %w", err)
		}
	}

	// Process models if present
	if r.HasModels() {
		// Check for circular dependencies
		cycles := DetectCircularDependencies(r.GetAllModels())
		if len(cycles) > 0 {
			fmt.Println("Warning: Circular dependencies detected in models:")
			for _, cycle := range cycles {
				fmt.Printf("  - %s\n", cycle.String())
			}
			fmt.Println("Note: Using TYPE_CHECKING imports to handle circular dependencies")
		}

		fmt.Println("Compiling models...")
		if err := CompileAllModels(config, r, writer); err != nil {
			return fmt.Errorf("failed to compile models: %w", err)
		}
	}

	// Process structures if present
	if r.HasStructures() {
		fmt.Println("Compiling structures...")
		if err := CompileAllStructures(config, r, writer); err != nil {
			return fmt.Errorf("failed to compile structures: %w", err)
		}
	}

	// Process entities if present
	if r.HasEntities() {
		// Entities depend on models
		if !r.HasModels() {
			return fmt.Errorf("entities compilation requires models to be compiled")
		}

		// Check for circular dependencies in entities
		entityCycles := DetectCircularDependencies(convertEntitiesToModels(r.GetAllEntities()))
		if len(entityCycles) > 0 {
			fmt.Println("Warning: Circular dependencies detected in entities:")
			for _, cycle := range entityCycles {
				fmt.Printf("  - %s\n", cycle.String())
			}
			fmt.Println("Note: Using TYPE_CHECKING imports to handle circular dependencies")
		}

		fmt.Println("Compiling entities...")
		if err := CompileAllEntities(config, r, writer); err != nil {
			return fmt.Errorf("failed to compile entities: %w", err)
		}
	}

	return nil
}

// convertEntitiesToModels converts entities to models for circular dependency checking
func convertEntitiesToModels(entities map[string]yaml.Entity) map[string]yaml.Model {
	models := make(map[string]yaml.Model)

	for name, entity := range entities {
		// Convert entity fields to model fields
		modelFields := make(map[string]yaml.ModelField)
		for fieldName, entityField := range entity.Fields {
			modelFields[fieldName] = yaml.ModelField{
				Type: yaml.ModelFieldType(entityField.Type),
			}
		}

		// Convert entity relations to model relations
		modelRelations := make(map[string]yaml.ModelRelation)
		for relName, entityRel := range entity.Related {
			modelRelations[relName] = yaml.ModelRelation{
				Type:    string(entityRel.Type),
				Aliased: entityRel.Aliased,
				For:     entityRel.For,
				Through: entityRel.Through,
			}
		}

		model := yaml.Model{
			Name:    entity.Name,
			Fields:  modelFields,
			Related: modelRelations,
		}
		models[name] = model
	}

	return models
}
