# Internal Package

The `internal/` directory contains application-specific packages that are not intended to be imported by external projects. It is organized by responsibility so that each part of the application has a clear place.

## Structure

```text
internal/
├── cache/
├── config/
├── contracts/
├── database/
├── encode/
├── handlers/
├── middleware/
├── models/
└── timesync/
```

## Packages

### `cache/`

Contains caching-related functionality.

Typical responsibilities:

* Store frequently accessed data
* Retrieve cached data
* Remove or invalidate cached data
* Manage cache configuration

### `config/`

Contains application configuration.

Typical responsibilities:

* Load environment variables
* Define application settings
* Provide configuration values to other packages
* Handle configuration defaults and validation

### `contracts/`

Contains interfaces and contracts used between different parts of the application.

Typical responsibilities:

* Define interfaces
* Define shared request/response contracts
* Reduce coupling between implementations

### `database/`

Contains database-related functionality.

Typical responsibilities:

* Database connection and initialization
* Database configuration
* Queries and persistence helpers
* Database lifecycle management

### `encode/`

Contains encoding and decoding functionality.

Typical responsibilities:

* Encode application data
* Decode incoming data
* Handle serialization formats
* Provide reusable encoding helpers

### `handlers/`

Contains application request handlers.

Typical responsibilities:

* Receive incoming requests
* Validate request data
* Call the appropriate application logic
* Build and return responses
* Handle request-level errors

### `middleware/`

Contains middleware used by the application.

Typical responsibilities:

* Authentication and authorization
* Request logging
* Error handling
* Request/response processing
* Other cross-cutting concerns

### `models/`

Contains application data models and domain structures.

Typical responsibilities:

* Define application entities
* Represent database records
* Define data structures shared across packages

### `timesync/`

Contains time synchronization functionality.

Typical responsibilities:

* Handle application time synchronization
* Provide consistent time-related utilities
* Manage communication with time synchronization services, if required

## Package Guidelines

When adding new code under `internal/`:

1. Keep each package focused on a single responsibility.
2. Avoid unnecessary dependencies between packages.
3. Keep reusable interfaces in `contracts/` when appropriate.
4. Keep HTTP/request-specific logic in `handlers/` and `middleware/`.
5. Keep database-specific logic inside `database/`.
6. Keep configuration-related logic inside `config/`.
7. Avoid exposing internal implementation details outside the application.

## Dependency Direction

A simple dependency flow should generally look like:

```text
handlers
   │
   ├── middleware
   ├── contracts
   ├── models
   └── database / cache
            │
            └── config
```

The exact dependency structure may differ depending on the application, but packages should avoid creating circular dependencies.

## Purpose

The goal of the `internal/` directory is to keep application-specific implementation details organized, maintainable, and isolated from the public API of the project.

