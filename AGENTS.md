# General instructions useful to everyone 

## Backend

The backend uses hexagonal architecture:
- `domain`: entities, value objects, business logic
- `application`: usecases, one handler for each command or query
- `adapters` driven adapters
- `entrypoints`: driving adapters

### Application and domain

- Structs have no public fields.
- Structs have constructors which validate their inputs.
- Structs are created through constructors.
- Constructors may be public or private.
- Constructors don't check for zero values of application structs.
- Returning a zero value of a struct is allowed together with a bool or an error.

### Enums

- Enums are structs with one private field of type string.
- Enum values are package-level variables.
- Enums don't have constructors.

### Tests

- Tests are in a `_test` package, for example `package example_test`.
