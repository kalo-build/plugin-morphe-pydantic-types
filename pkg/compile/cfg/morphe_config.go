package cfg

import "fmt"

// MorpheConfig contains configuration for all Morphe type categories
// This is a simplified version without language-specific details
type MorpheConfig struct {
	// Configuration for different type categories
	Enums      EnumConfig      `json:"enums,omitempty"`
	Models     ModelConfig     `json:"models,omitempty"`
	Structures StructureConfig `json:"structures,omitempty"`
	Entities   EntityConfig    `json:"entities,omitempty"`
}

// EnumConfig contains configuration specific to enum generation
type EnumConfig struct {
	// GenerateHelpers controls whether to generate helper methods
	GenerateStrMethod bool `json:"generateStrMethod,omitempty"`
	// UseStrEnum uses StrEnum for string-based enums (Python 3.11+)
	UseStrEnum bool `json:"useStrEnum,omitempty"`
}

// ModelConfig contains configuration specific to model generation
type ModelConfig struct {
	// UseField controls whether to use Pydantic Field for model fields
	UseField bool `json:"useField,omitempty"`
	// GenerateExamples adds example values in Field definitions
	GenerateExamples bool `json:"generateExamples,omitempty"`
	// UseValidators generates Pydantic validators for common patterns
	UseValidators bool `json:"useValidators,omitempty"`
}

// StructureConfig contains configuration specific to structure generation
type StructureConfig struct {
	// UseDataclass generates Python dataclasses instead of Pydantic models
	UseDataclass bool `json:"useDataclass,omitempty"`
	// GenerateSlots adds __slots__ for memory efficiency
	GenerateSlots bool `json:"generateSlots,omitempty"`
}

// EntityConfig contains configuration specific to entity generation
type EntityConfig struct {
	// GenerateRepository generates repository pattern methods
	GenerateRepository bool `json:"generateRepository,omitempty"`
	// LazyLoadingStyle controls lazy loading implementation
	LazyLoadingStyle string `json:"lazyLoadingStyle,omitempty"` // "async", "sync", "property"
	// IncludeValidation adds validation methods
	IncludeValidation bool `json:"includeValidation,omitempty"`
}

// Validate checks if the configuration is valid
func (config MorpheConfig) Validate() error {
	// Validate entity lazy loading style
	if config.Entities.LazyLoadingStyle != "" {
		validStyles := map[string]bool{
			"async":    true,
			"sync":     true,
			"property": true,
		}
		if !validStyles[config.Entities.LazyLoadingStyle] {
			return fmt.Errorf("invalid lazy loading style: %s (must be 'async', 'sync', or 'property')",
				config.Entities.LazyLoadingStyle)
		}
	}

	// No other validations needed as all other options are boolean flags
	return nil
}
