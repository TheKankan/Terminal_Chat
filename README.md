![CI](https://github.com/TheKankan/Terminal_Chat/actions/workflows/ci.yml/badge.svg)

# Terminal_Chat

A terminal-based chat application allowing multiple users to communicate through a client-server architecture.

## Motivation

This project was built to explore how to design a secure, self-hosted chat system without relying on third-party platforms.
It focuses on building a backend from scratch, handling environment configuration, database persistence, and client-server communication in Go.

## Quick Start

### Prerequisites
- Go Version 1.25+ installed, you can download it [here](https://go.dev/doc/install)
- PostgreSQL

### Running

For the program to work you will need to configure a server, the clients will then be able to connect to your server.

- Clone the repo 
- Replace the variables in .env.example with your values and rename it .env (you will need to create your postgres database)
- Migrate up everything in sql/schema
- Type `./server` in your terminal

Congratulation, the server should now be running !

- You can now launch any number of clients in other terminals with `./client`

## Contributing

- Clone the repo 
- Replace the variables in .env.example with your values and rename it .env (you will need to create your postgres database)
- Migrate up everything in sql/schema
- Use `go run ./cmd/server` to run the server and then `go run ./cmd/client` in another terminal for each client

You can open a pull request to the `main` branch to add new features or fix issues.