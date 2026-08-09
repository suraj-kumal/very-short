# `/cmd` Directory in Go

The `/cmd` directory is a common Go project convention used to organize the **executable applications** of a project.

It is not required by the Go language.

## Purpose

The main purpose of `/cmd` is to keep **application entry points** separate from the rest of the project's code.

For example:

```text
project/
├── cmd/
│   ├── api/
│   │   └── main.go
│   ├── worker/
│   │   └── main.go
│   └── migrate/
│       └── main.go
│
├── internal/
├── pkg/
└── go.mod
```

Each directory inside `/cmd` represents a separate executable application.

## Why `/cmd`?

A project may contain more than one executable.

For example:

```text
cmd/
├── api/
│   └── main.go       # API server
│
├── worker/
│   └── main.go       # Background worker
│
└── migrate/
    └── main.go       # Database migration tool
```

This gives the project a clear structure:

```text
cmd/api      → API application
cmd/worker   → Worker application
cmd/migrate  → Migration application
```

Each one can be run or built independently.

```bash
go run ./cmd/api
go run ./cmd/worker
go run ./cmd/migrate
```

Or built independently:

```bash
go build ./cmd/api
go build ./cmd/worker
go build ./cmd/migrate
```

## `/cmd` Does Not Contain Business Logic

The `/cmd` directory should generally contain the **entry point and startup configuration** for an executable.

For example:

```go
package main

func main() {
    // Initialize dependencies
    // Configure the application
    // Start the server
}
```

The actual application logic should normally live in other packages.

Conceptually:

```text
cmd/
    ↓
Entry point
    ↓
Application setup
    ↓
Internal packages
    ↓
Business logic
```

## Why Not Put Everything in `/cmd`?

Consider:

```text
cmd/
└── api/
    └── main.go
```

`main.go` should not become a huge file containing:

```text
HTTP handlers
Business logic
Database queries
Authentication
Validation
Email logic
etc.
```

Instead, `main.go` should primarily **assemble the application**.

For example:

```go
func main() {
    db := database.New()

    repo := repository.New(db)

    service := service.New(repo)

    handler := handler.New(service)

    server := server.New(handler)

    server.Start()
}
```

The `/cmd` application acts as the **composition root**: it connects the different parts of the application together.

## Naming Directories Inside `/cmd`

Names should describe the executable.

Common examples:

```text
cmd/
├── api/
├── worker/
├── migrate/
├── cli/
└── admin/
```

For example:

```text
cmd/api/
```

means:

> The executable that runs the API.

```text
cmd/worker/
```

means:

> The executable that runs the background worker.

```text
cmd/migrate/
```

means:

> The executable that performs database migrations.

## `/cmd` Is a Convention

Go does **not** require a `/cmd` directory.

This is perfectly valid:

```text
project/
├── main.go
└── go.mod
```

You could also have:

```text
project/
├── api/
│   └── main.go
└── go.mod
```

The `/cmd` convention becomes useful when the project has multiple executable applications or when you want a clear separation between **entry points** and **application implementation**.

## Simple Mental Model

Think of `/cmd` as:

```text
/cmd = "What can I run?"
```

For example:

```text
cmd/
├── api/       → Run the API
├── worker/    → Run the worker
└── migrate/   → Run migrations
```

Each subdirectory is an executable entry point.

## Summary

* `/cmd` is a **convention**, not a Go requirement.
* It contains the project's **executable applications**.
* Each subdirectory usually represents a separate executable.
* `main.go` is commonly placed inside each executable's directory.
* Keep the entry-point code focused on **startup and dependency wiring**.
* Keep business logic in separate packages.

The key idea:

```text
/cmd
  └── contains the programs you can run
```

