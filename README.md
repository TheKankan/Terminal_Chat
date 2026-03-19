![CI](https://github.com/TheKankan/Terminal_Chat/actions/workflows/ci.yml/badge.svg)

# Terminal Chat

A terminal-based chat application allowing multiple users to communicate through a secure client-server architecture.

## Motivation

This project was built to explore how to design a secure, self-hosted chat system without relying on third-party platforms.
It focuses on building a backend from scratch, handling environment configuration, database persistence, and client-server communication in Go.

## Features

- 🔐 Authentication with hashed passwords (argon2id) and JWT tokens
- 💬 Real-time messaging via WebSocket
- 🗂️ Chat history stored in PostgreSQL
- 🖥️ Terminal UI with color-coded messages (bubbletea)
- 🐳 One-command setup with Docker

## Quick Start (Docker)

### Prerequisites

- [Docker](https://www.docker.com/get-started) installed

### Running

```bash
# 1. Clone the repo
git clone https://github.com/TheKankan/Terminal_Chat.git
cd Terminal_Chat

# 2. Edit .env.example : rename it .env and change the values inside it

# 3. Start the server and database
docker compose up
```

The server is now running on port `8080` !

To connect as a client, in a new terminal :

```bash
go run ./cmd/client
```

You can open as many clients as you want in separate terminals.

## Manual Setup (without Docker)

### Prerequisites

- Go 1.25+
- PostgreSQL
- [goose](https://github.com/pressly/goose) for migrations

### Running

```bash
# 1. Clone the repo
git clone https://github.com/TheKankan/Terminal_Chat.git
cd Terminal_Chat

# 2. Edit .env.example : rename it .env and change the values inside it

# 3. Run migrations
goose -dir sql/schema postgres "$DB_URL" up

# 4. Start the server
go run ./cmd/server

# 5. In separate terminals, start clients
go run ./cmd/client
```

## Contributing

```bash
# Clone the repo and set up your .env (see Manual Setup above)

# Run tests
go test ./...

# Start the server
go run ./cmd/server

# Start a client
go run ./cmd/client
```

Open a pull request to the `main` branch to add new features or fix issues.

## Tech Stack

- **Go** — server, client, and TUI
- **PostgreSQL** — persistent storage
- **WebSocket** — real-time communication ([gorilla/websocket](https://github.com/gorilla/websocket))
- **JWT** — authentication ([golang-jwt](https://github.com/golang-jwt/jwt))
- **argon2id** — password hashing ([alexedwards/argon2id](https://github.com/alexedwards/argon2id))
- **sqlc** — type-safe SQL ([sqlc](https://sqlc.dev))
- **bubbletea** — terminal UI ([charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea))
- **Docker** — containerization
