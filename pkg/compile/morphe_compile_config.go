package compile

import (
	"path"

	rcfg "github.com/kalo-build/morphe-go/pkg/registry/cfg"
	"github.com/kalo-build/plugin-morphe-pydantic-types/pkg/compile/cfg"
)

// MorpheCompileConfig contains all configuration for compiling Morphe to the target format
type MorpheCompileConfig struct {
	// Registry loading configuration
	rcfg.MorpheLoadRegistryConfig

	// Output path for generated files
	OutputPath string

	// Format-specific configuration
	FormatConfig PydanticConfig

	// Type-specific configuration
	MorpheConfig cfg.MorpheConfig
}

// PydanticConfig contains Pydantic-specific configuration options
type PydanticConfig struct {
	// Pydantic-specific options
	PydanticV2    bool   `json:"pydanticV2"`    // Use Pydantic v2 syntax (default: true)
	AddTypeHints  bool   `json:"addTypeHints"`  // Add type hints (default: true)
	GenerateInit  bool   `json:"generateInit"`  // Generate __init__.py files (default: true)
	IndentSize    int    `json:"indentSize"`    // Number of spaces for indent (default: 4)
	PythonVersion string `json:"pythonVersion"` // Target Python version (default: "3.8")
}

// DefaultMorpheCompileConfig creates a default configuration
func DefaultMorpheCompileConfig(
	yamlRegistryPath string,
	baseOutputDirPath string,
) MorpheCompileConfig {
	return MorpheCompileConfig{
		MorpheLoadRegistryConfig: rcfg.MorpheLoadRegistryConfig{
			RegistryEnumsDirPath:      path.Join(yamlRegistryPath, "enums"),
			RegistryModelsDirPath:     path.Join(yamlRegistryPath, "models"),
			RegistryStructuresDirPath: path.Join(yamlRegistryPath, "structures"),
			RegistryEntitiesDirPath:   path.Join(yamlRegistryPath, "entities"),
		},
		OutputPath: baseOutputDirPath,
		FormatConfig: PydanticConfig{
			PydanticV2:    true,
			AddTypeHints:  true,
			GenerateInit:  true,
			IndentSize:    4,
			PythonVersion: "3.8",
		},
	}
}

// Validate checks if the configuration is valid
func (config MorpheCompileConfig) Validate() error {
	// Validate registry paths
	if err := config.MorpheLoadRegistryConfig.Validate(); err != nil {
		return err
	}

	// TODO: Add format-specific validation
	// Examples:
	// - Check if package prefix is valid
	// - Verify indent size is positive
	// - Ensure file extension starts with "."

	return nil
}
