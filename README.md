# Morphe Pydantic Types Plugin

A Morphe compilation plugin that generates Python code with Pydantic models for data validation and serialization.

## Features

- ✅ Generates Python 3.8+ compatible code
- ✅ Uses Pydantic v2 for data validation
- ✅ Full type hints support
- ✅ Automatic `__init__.py` generation
- ✅ Handles enums, models, structures, and entities
- ✅ Relationship support with lazy loading patterns
- ✅ **Polymorphic relationships** (ForOnePoly, HasManyPoly, etc.)
- ✅ **Aliasing support** for custom relationship naming
- ✅ Integration tests with ground truth validation

## Generated Output Example

### Enum
```python
class Nationality(Enum):
    """Nationality enumeration."""
    D_E = "German"
    F_R = "French"
    U_S = "American"
```

### Model (Pydantic)
```python
class Person(BaseModel):
    """Person model."""
    first_name: str
    id: int
    last_name: str
    nationality: Nationality
    company_id: Optional[str] = None
```

### Entity
```python
class Company(BaseModel):
    """Company entity."""
    id: int  # primary identifier
    name: str
    tax_id: str
    persons: List[Person] = None
    
    async def load_persons(self) -> List['Person']:
        """Load related Person entities."""
        # TODO: Implement lazy loading
        return []
```

### Polymorphic Model
```python
class Comment(BaseModel):
    """Comment model with polymorphic relationship."""
    content: str
    id: int
    commentable_type: Optional[str] = None
    commentable_id: Optional[str] = None
    commentable: Optional[Union['Person', 'Company']] = None
```

## Usage

```bash
# Build the plugin
go build ./cmd/plugin

# Generate Python code
./plugin '{"inputPath":"./morphe","outputPath":"./output","verbose":true}'
```

## Configuration

The plugin supports comprehensive Python-specific and type-specific options:

```json
{
  "inputPath": "./morphe",
  "outputPath": "./output",
  "config": {
    // Python-specific settings
    "pythonVersion": "3.11",
    "usePydantic": true,
    "pydanticV2": true,
    "addTypeHints": true,
    "generateInit": true,
    "indentSize": 4,
    
    // Type-specific configurations
    "enums": {
      "generateStrMethod": true,
      "useStrEnum": false
    },
    "models": {
      "useField": true,
      "generateExamples": false,
      "useValidators": false
    },
    "structures": {
      "useDataclass": false,
      "generateSlots": false
    },
    "entities": {
      "generateRepository": false,
      "lazyLoadingStyle": "async",
      "includeValidation": false
    }
  }
}
```

See [KALO_CONFIG_EXAMPLE.md](KALO_CONFIG_EXAMPLE.md) for detailed configuration options and kalo.yaml integration.

## Testing

### Run Integration Tests
```bash
go test ./pkg/compile -v
```

### Inspect Generated Code
```bash
ls -la output/
```

The plugin includes:
- **Ground truth tests** that ensure output matches expected files
- **Comprehensive test coverage** for all generated code patterns

## Project Structure

```
plugin-morphe-py-types/
├── cmd/plugin/          # Entry point
├── pkg/
│   ├── compile/         # Core compilation logic
│   ├── formatdef/       # Python type definitions
│   └── typemap/         # Morphe → Python type mappings
├── testdata/            # Test schemas and ground truth
│   ├── registry/        # Input test schemas
│   └── ground-truth/    # Expected outputs
└── output/             # Generated Python code
```

## Known Limitations

- Enum imports in models are tracked but require the enums to be accessible
- Generated code uses relative imports (standard for Python packages)
- Entity relationship loading is stubbed (requires actual implementation)

## License

Same as other Morphe plugins.