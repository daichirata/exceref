# exceref

exceref is a CLI that reads Excel sheets with reference definitions, resolves them, and exports data or generates code.

## Features
- Resolve references and export data (csv/json/yaml)
- Update reference data and data validations
- Code generation with templates (go/csharp/generic)
- Metadata export to YAML

## Data sheet format
The first three rows are treated as headers:
- Row 1: Type
- Row 2: Column name
- Row 3: Description (optional)

Subsequent rows are data. Columns with an empty Name or Type are skipped for export.

### Supported types
- string
- int
- float
- bool
- datetime
- date
- unixtime
- ref

## Reference definition sheet (_references)
Reference definitions live in the `_references` sheet with these column names:
- sheet
- column
- reference_file
- reference_sheet
- reference_key
- reference_value
- reference_name

If `reference_value` is empty, it is treated as a polymorphic reference.

## Usage
```
exceref export -o out -f csv -p prefix_ path/to/book.xlsx
exceref export -o out -f json path/to/book.xlsx
exceref export -o out -f yaml path/to/book.xlsx

exceref generate -o out -l go -t path/to/template.tmpl path/to/book.xlsx
exceref generate -o out -l csharp -t path/to/template.tmpl path/to/book.xlsx

exceref update path/to/book.xlsx

exceref meta export -o out path/to/book.xlsx
```

## Template data
Templates receive:
- Name: singularized, Camel/Pascalized sheet name
- Fields: []Field with Name, ColumnName, Type

Template functions `camelize` and `singularize` are available.
