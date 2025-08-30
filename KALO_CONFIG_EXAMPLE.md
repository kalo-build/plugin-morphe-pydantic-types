# Kalo Configuration Example

This shows how to configure the Python plugin in your `kalo.yaml` file:

```yaml
# kalo.yaml
stores:
  KA_MO_YAML_PATH:
    path: ./morphe/registry
  KA_MO_PY_PATH:
    path: ./python

stages:
  - name: py-types
    parallel: false
    plugins:
      - name: "@kalo-build/plugin-morphe-pydantic-types"
        version: 1.0.0
        input:
          store: KA_MO_YAML_PATH
        output:
          store: KA_MO_PY_PATH
        config:
          # Python-specific settings
          pythonVersion: "3.11"
          usePydantic: true
          pydanticV2: true
          addTypeHints: true
          generateInit: true
          indentSize: 4
          
          # Type-specific configurations
          enums:
            generateStrMethod: true
            useStrEnum: false  # Use StrEnum for Python 3.11+
          
          models:
            useField: true  # Use Pydantic Field for all fields
            generateExamples: false
            useValidators: false
          
          structures:
            useDataclass: false  # Use dataclasses instead of Pydantic
            generateSlots: false
          
          entities:
            generateRepository: false
            lazyLoadingStyle: "async"  # Options: "async", "sync", "property"
            includeValidation: false
```

## Configuration Options

### Global Python Settings

- `pythonVersion`: Target Python version (default: "3.8")
- `usePydantic`: Use Pydantic for models (default: true)
- `pydanticV2`: Use Pydantic v2 syntax (default: true)
- `addTypeHints`: Add type hints (default: true)
- `generateInit`: Generate `__init__.py` files (default: true)
- `indentSize`: Spaces per indent level (default: 4)

### Enum Configuration

- `generateStrMethod`: Add `__str__` method to enums
- `useStrEnum`: Use `StrEnum` for string enums (Python 3.11+)

### Model Configuration

- `useField`: Use Pydantic `Field` for model fields
- `generateExamples`: Add example values in Field definitions
- `useValidators`: Generate Pydantic validators

### Structure Configuration

- `useDataclass`: Generate Python dataclasses instead of Pydantic models
- `generateSlots`: Add `__slots__` for memory efficiency

### Entity Configuration

- `generateRepository`: Generate repository pattern methods
- `lazyLoadingStyle`: Style for lazy loading ("async", "sync", "property")
- `includeValidation`: Add validation methods

## Minimal Configuration

If you're happy with the defaults, you can use a minimal configuration:

```yaml
config: {}  # Uses all defaults
```

## Advanced Example

For a project using modern Python with custom requirements:

```yaml
config:
  pythonVersion: "3.12"
  pydanticV2: true
  indentSize: 2
  enums:
    useStrEnum: true  # Use StrEnum for better IDE support
  models:
    useField: true
    useValidators: true  # Generate email/phone validators
  entities:
    lazyLoadingStyle: "property"  # Use @property for lazy loading
```
