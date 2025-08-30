# Test Data for Python Morphe Plugin

This directory contains test data and ground truth files for the Python Pydantic plugin.

## Directory Structure

```
testdata/
├── registry/          # Input Morphe schema files
│   └── minimal/      # Minimal test case
│       ├── entities/
│       ├── enums/
│       ├── models/
│       └── structures/
└── ground-truth/     # Expected output files
    └── compile-minimal/
        ├── entities/
        ├── enums/
        ├── models/
        └── structures/
```

## Running Tests

### 1. Unit/Integration Tests

Run the Go tests that compare output against ground truth:

```bash
go test ./pkg/compile -v
```

This will:
- Generate Python code from the test schemas
- Compare against ground truth files
- Ensure all expected files are created

### 2. Manual Validation

You can manually inspect the generated code in the `output/` directory to verify it meets your expectations.

## Ground Truth Files

The ground truth files in `ground-truth/compile-minimal/` represent the expected output for the minimal test case. These files are:

- **Generated automatically** from the current plugin implementation
- **Validated** to be syntactically correct Python
- **Include all features**: enums, models, structures, and entities
- **Use Pydantic v2** with type hints

## Updating Ground Truth

If you make intentional changes to the output format:

1. Regenerate the output:
   ```bash
   ./plugin.exe '{"inputPath":"./testdata/registry/minimal","outputPath":"./output"}'
   ```

2. Inspect the output:
   ```bash
   ls -la output/
   ```

3. Update ground truth:
   ```bash
   rm -rf testdata/ground-truth/compile-minimal
   cp -r output testdata/ground-truth/compile-minimal
   ```

4. Run tests to ensure they pass:
   ```bash
   go test ./pkg/compile -v
   ```

## Adding New Test Cases

To add a new test case:

1. Create a new schema set in `testdata/registry/new-case/`
2. Generate the expected output
3. Create ground truth in `testdata/ground-truth/compile-new-case/`
4. Add a new test method in `compile_test.go`

## Notes

- The generated code uses relative imports, which is standard for Python packages
- Each directory includes an `__init__.py` file for proper package structure
- Pydantic models include validation and configuration
- Entities handle relationships and lazy loading patterns
