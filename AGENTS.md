# oapibase

## Overview

This project is a general purpose API template.

- Tech Stack: Go, OpenAPI, Postgres

## Development

- Use `task` to run commands, see `Taskfile.yml` for available tasks.
- See the `/migrations` folder for the database schema. Use tasks `migrate-up`, `migrate-down`, and `migrate-add` to manage migrations. Default to postgres.
- OpenAPI spec is created declaratively when mounting routes in `api/router.go`. Uses github.com/oaswrap/spec.


## Code Structure

- Use standard net/http handlers.
- Define api handlers in the `/api` folder.
- Domain specific logic can go in individual folders in `/api`.
