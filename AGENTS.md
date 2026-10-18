# General instructions useful to everyone 

## Backend

The backend uses hexagonal architecture:
- `domain`: entities, value objects, business logic
- `application`: usecases, one handler for each command or query
- `adapters` driven adapters
- `entrypoints`: driving adapters

### Domain layer

- Domain types are structs as well as functions and interfaces.
- Domain structs have no public fields. An embedded exported struct is allowed.
- Domain structs have constructors which validate their inputs.
- Domain structs are created through constructors.
- Constructors may be public or private.
- Constructors don't check for zero values of domain structs.
- Enums are structs with one private field of type string.
- Enums values are package-level variables.

### Tests

- Tests are in a `_test` package, for example `package example_test`.
