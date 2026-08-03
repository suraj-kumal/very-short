# URL Shortner

## folder/file structure
j main.go : Entry point of the application; initializes the server, database, and routes.
- handlers.go : Contains HTTP request handlers for shortening URLs and redirecting users.
- models.go : Defines the application's data models and structures.
- db.go : Manages the database connection and database operations.
- config.go : Stores and loads application configuration settings.
- templates : Contains HTML templates used to render web pages.
- static : Stores static assets such as CSS, JavaScript, and images.
- go.mod : Defines the Go module and manages project dependencies.
- go.sum : Stores checksums to ensure dependency integrity.
