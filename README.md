# sano.redirectme.net

## Project Overview

`sano.redirectme.net` is a lightweight Go-powered URL shortener built for fast redirects and low-latency delivery. The application provides a web interface for shortening a long URL, stores shortened links in a MySQL database, and serves redirect requests through an in-memory cache for improved performance.

Key features:

- HTTP web UI for submitting URLs and receiving a short link
- Short code generation using base62 encoding and a secret multiplier
- MySQL-backed persistent storage for URL records
- In-memory LRU cache to speed up redirect lookups
- Expiration and 404/410 error handling
- Background synchronization of access timestamps from cache to database

## Installation Instructions

### Prerequisites

- Go 1.26 or later
- MySQL database server
- Git

### Clone the repository

```bash
git clone https://github.com/suraj-kumal/very-short.git
cd very-short
```

### Install dependencies

```bash
go mod download
```

### Configure environment variables

Create a `.env` file in the project root with the following values:

```env
DATABASE_URL=user:password@tcp(localhost:3306)/your_database_name
PORT=8080
SITE_URL=http://localhost:8080
MIX_MULTIPLIER_SECRET=12345
```

- `DATABASE_URL`: MySQL connection string
- `PORT`: port number for the application to listen on
- `SITE_URL`: full public URL used to build shortened links
- `MIX_MULTIPLIER_SECRET`: numeric secret used during short code generation

### Database schema example

The application expects a `url_data` table with at least the following columns:

```sql
CREATE TABLE url_data (
  id INT AUTO_INCREMENT PRIMARY KEY,
  url TEXT NOT NULL,
  hash VARCHAR(255),
  expire_at DATETIME NOT NULL DEFAULT '9999-12-31 23:59:59',
  LastAccessTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### Build and run

```bash
go build -o very-short
./very-short
```

The server will start and listen on the port configured in `.env`.

## Usage Guide

### Web interface

Open your browser at:

```text
http://localhost:8080
```

Use the form to submit a long URL. The page will return a shortened URL that can be copied and shared.

### API endpoint

Shorten a URL using a `POST` request to `/shorten`.

```bash
curl -X POST \
  -d "url=https://example.com/very/long/path" \
  http://localhost:8080/shorten
```

The server responds with HTML content containing the shortened URL.

### Redirect behavior

Visit the generated short code URL in your browser, for example:

```text
http://localhost:8080/abc123
```

The service checks the cache first, and if the link is still valid it redirects to the original URL.

### Important notes

- URLs must begin with `http://` or `https://`
- Expired URLs return a `410 Gone` page
- Missing or invalid short codes return a `404 Not Found` page
- Static assets are served from the `static/` directory

## Contributing Guidelines

Contributions are welcome. If you want to improve the repository, follow these steps:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/name`
3. Make your code changes
4. Format Go code with `gofmt` before submitting
5. Commit and push your branch
6. Open a pull request with a clear description of your changes

### Code style

- Use idiomatic Go patterns
- Keep handlers, cache, and database logic separated
- Use `gofmt` for formatting

### Testing

This repository does not currently include automated tests. If you add tests, make sure they cover the added behavior and run with:

```bash
go test ./...
```

### Issues and pull requests

- Open an issue for bugs or feature requests
- Reference the issue in your pull request
- Keep pull requests small and focused

## License Information

This repository does not currently include a dedicated `LICENSE` file. Add a license file to clearly define the terms under which the project may be used, modified, and distributed.

## Contact Information

For support or questions:

- Open an issue in this repository
- If this repository is hosted on GitHub, use the repository issue tracker for feedback

If you want to provide direct contact details, add them to this section after the repository owner decides on the preferred channel.
