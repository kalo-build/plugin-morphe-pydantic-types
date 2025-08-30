package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kalo-build/plugin-morphe-pydantic-types/pkg/compile"
	"github.com/kalo-build/plugin-morphe-pydantic-types/pkg/compile/cfg"
)

// CompileConfig represents the configuration passed to the plugin
type CompileConfig struct {
	InputPath  string       `json:"inputPath"`
	OutputPath string       `json:"outputPath"`
	Config     PluginConfig `json:"config,omitempty"`
	Verbose    bool         `json:"verbose,omitempty"`
}

// PluginConfig represents the Pydantic-specific configuration
type PluginConfig struct {
	// Pydantic-specific settings
	PythonVersion string `json:"pythonVersion,omitempty"`
	PydanticV2    *bool  `json:"pydanticV2,omitempty"`
	AddTypeHints  *bool  `json:"addTypeHints,omitempty"`
	GenerateInit  *bool  `json:"generateInit,omitempty"`
	IndentSize    *int   `json:"indentSize,omitempty"`

	// Type-specific configurations
	Enums      cfg.EnumConfig      `json:"enums,omitempty"`
	Models     cfg.ModelConfig     `json:"models,omitempty"`
	Structures cfg.StructureConfig `json:"structures,omitempty"`
	Entities   cfg.EntityConfig    `json:"entities,omitempty"`
}

// Exit codes
const (
	ExitSuccess         = 0
	ExitCompileFailed   = 1
	ExitMissingConfig   = 3
	ExitInvalidConfig   = 4
	ExitInputPathError  = 12
	ExitOutputPathError = 13
)

// logInfo prints info messages only when verbose mode is enabled
func logInfo(verbose bool, format string, args ...interface{}) {
	if verbose {
		fmt.Fprintf(os.Stdout, format+"\n", args...)
	}
}

func main() {
	// Check command line arguments
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: plugin-morphe-pydantic-types <config>")
		fmt.Fprintln(os.Stderr, "  config: JSON string with inputPath, outputPath, and optional config parameters")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Example:")
		fmt.Fprintln(os.Stderr, `  plugin-morphe-pydantic-types '{"inputPath":"./morphe","outputPath":"./output","verbose":true}'`)
		os.Exit(ExitMissingConfig)
	}

	// Parse configuration
	rawConfig := os.Args[1]
	var compileConfig CompileConfig
	if err := json.Unmarshal([]byte(rawConfig), &compileConfig); err != nil {
		fmt.Fprintln(os.Stderr, "Error parsing config JSON:", err)
		fmt.Fprintln(os.Stderr, "Expected format: {\"inputPath\":\"...\",\"outputPath\":\"...\",\"config\":{...},\"verbose\":false}")
		os.Exit(ExitInvalidConfig)
	}

	// Validate required fields
	if compileConfig.InputPath == "" {
		fmt.Fprintln(os.Stderr, "Error: inputPath is required")
		os.Exit(ExitInputPathError)
	}

	if compileConfig.OutputPath == "" {
		fmt.Fprintln(os.Stderr, "Error: outputPath is required")
		os.Exit(ExitOutputPathError)
	}

	// Convert to absolute paths
	inputAbs, err := filepath.Abs(compileConfig.InputPath)
	if err == nil {
		compileConfig.InputPath = inputAbs
	}

	outputAbs, err := filepath.Abs(compileConfig.OutputPath)
	if err == nil {
		compileConfig.OutputPath = outputAbs
	}

	logInfo(compileConfig.Verbose, "Processing Morphe registry from: '%s'", compileConfig.InputPath)
	logInfo(compileConfig.Verbose, "Output Pydantic types to: '%s'", compileConfig.OutputPath)

	// Initialize the compile configuration
	logInfo(compileConfig.Verbose, "Initializing compile configuration...")
	morpheConfig := compile.DefaultMorpheCompileConfig(
		compileConfig.InputPath,
		compileConfig.OutputPath,
	)

	// Apply configuration from compileConfig.Config
	// Python version
	if compileConfig.Config.PythonVersion != "" {
		morpheConfig.FormatConfig.PythonVersion = compileConfig.Config.PythonVersion
		logInfo(compileConfig.Verbose, "Setting Python version to: %s", compileConfig.Config.PythonVersion)
	}

	// Pydantic settings
	if compileConfig.Config.PydanticV2 != nil {
		morpheConfig.FormatConfig.PydanticV2 = *compileConfig.Config.PydanticV2
		logInfo(compileConfig.Verbose, "Use Pydantic v2: %v", *compileConfig.Config.PydanticV2)
	}

	// Type hints
	if compileConfig.Config.AddTypeHints != nil {
		morpheConfig.FormatConfig.AddTypeHints = *compileConfig.Config.AddTypeHints
		logInfo(compileConfig.Verbose, "Add type hints: %v", *compileConfig.Config.AddTypeHints)
	}

	// Init files
	if compileConfig.Config.GenerateInit != nil {
		morpheConfig.FormatConfig.GenerateInit = *compileConfig.Config.GenerateInit
		logInfo(compileConfig.Verbose, "Generate __init__.py: %v", *compileConfig.Config.GenerateInit)
	}

	// Indentation
	if compileConfig.Config.IndentSize != nil {
		morpheConfig.FormatConfig.IndentSize = *compileConfig.Config.IndentSize
		logInfo(compileConfig.Verbose, "Indent size: %d", *compileConfig.Config.IndentSize)
	}

	// Apply type-specific configurations
	morpheConfig.MorpheConfig.Enums = compileConfig.Config.Enums
	morpheConfig.MorpheConfig.Models = compileConfig.Config.Models
	morpheConfig.MorpheConfig.Structures = compileConfig.Config.Structures
	morpheConfig.MorpheConfig.Entities = compileConfig.Config.Entities

	// Log type-specific configs if verbose
	if compileConfig.Verbose {
		if compileConfig.Config.Models.UseField {
			logInfo(true, "Models use Field: true")
		}
		if compileConfig.Config.Enums.GenerateStrMethod {
			logInfo(true, "Enums generate __str__: true")
		}
		if compileConfig.Config.Entities.LazyLoadingStyle != "" {
			logInfo(true, "Entity lazy loading style: %s", compileConfig.Config.Entities.LazyLoadingStyle)
		}
	}

	// Validate configuration
	if err := morpheConfig.Validate(); err != nil {
		fmt.Fprintln(os.Stderr, "Invalid configuration:", err)
		os.Exit(ExitInvalidConfig)
	}

	// Run compilation
	logInfo(compileConfig.Verbose, "Starting compilation process...")
	if err := compile.MorpheToPydantic(morpheConfig); err != nil {
		fmt.Fprintln(os.Stderr, "Compilation failed:", err)
		os.Exit(ExitCompileFailed)
	}

	logInfo(compileConfig.Verbose, "Compilation completed successfully")
	os.Exit(ExitSuccess)
}
