# Issues - Claude Code Project Guide

## Project Overview

AI-powered issue management system. When an issue is opened, Claude Code automatically executes the task. Built with Go backend, React web, React Native mobile, and PostgreSQL.

## Tech Stack

- **Backend**: Go (echo, pgx/sqlx, golang-jwt, x/oauth2)
- **Frontend Web**: React + TypeScript + Vite + Tailwind CSS
- **Frontend Mobile**: React Native + Expo
- **Database**: PostgreSQL 16
- **AI Runtime**: Claude Code CLI (subprocess execution)
- **Auth**: Google / GitHub OAuth + JWT

## Project Structure

```
cmd/server/            # Application entry point
internal/
  config/              # Environment-based configuration
  domain/              # Entities and domain errors
  handler/             # HTTP handlers, middleware, response helpers
  service/             # Business logic, AI runner, worker pool
  repository/          # PostgreSQL data access (sqlx)
migrations/            # golang-migrate SQL files
api/                   # OpenAPI spec
web/                   # React web app
mobile/                # React Native app
.claude/
  commands/            # Slash commands (/plan, /tdd, /code-review, etc.)
  rules/common/        # Universal coding rules
  rules/golang/        # Go-specific rules
agents/                # Specialized subagents
skills/                # Domain knowledge and patterns
```

## Agents

Available agents in `agents/`:
- `planner.md` - Feature implementation planning
- `architect.md` - System design decisions
- `tdd-guide.md` - Test-driven development
- `code-reviewer.md` - Quality and security review
- `security-reviewer.md` - Vulnerability analysis
- `build-error-resolver.md` - Build error resolution
- `go-reviewer.md` - Go code review
- `go-build-resolver.md` - Go build error resolution
- `database-reviewer.md` - Database design review

## Key Commands

- `/plan` - Implementation planning
- `/tdd` - Test-driven development workflow
- `/code-review` - Code quality review
- `/build-fix` - Fix build errors
- `/go-review` - Go code review
- `/go-test` - Go TDD workflow
- `/go-build` - Fix Go build errors
- `/verify` - Run verification loop
- `/checkpoint` - Save verification state

## Development Commands (Taskfile)

```bash
task build          # Build server binary
task run            # Run server
task test           # Run tests with race detection
task test:verbose   # Verbose test output
task lint           # Run golangci-lint
task fmt            # Format code (gofmt + goimports)
task migrate:up     # Run migrations
task migrate:down   # Rollback migrations
task db:up          # Start PostgreSQL (Docker)
task db:down        # Stop PostgreSQL
task dev            # Start DB + run server
```

## Development Guidelines

- Follow rules in `.claude/rules/common/` and `.claude/rules/golang/`
- **Immutability**: Always create new objects, never mutate existing ones
- **TDD**: Write tests first (RED → GREEN → REFACTOR), 80%+ coverage
- **File size**: < 800 lines per file, < 50 lines per function
- **Error handling**: Wrap with context (`fmt.Errorf("context: %w", err)`)
- **Interfaces**: Define where consumed (in `service/`), not where implemented
- **API responses**: Use envelope format (`{ "data": ..., "error": ... }`)
- **Pagination**: Cursor-based for all list endpoints
- **Commits**: `<type>: <description>` (feat, fix, refactor, docs, test, chore)

## Key Architectural Decisions

- **Repository Pattern**: Abstract data access behind interfaces for testability
- **Service Layer**: Business logic separated from HTTP handling
- **PostgreSQL Job Queue**: `FOR UPDATE SKIP LOCKED` for AI job processing (no Redis needed)
- **JWT Auth**: Access token (15min) + Refresh token (7 days), no per-user permissions
- **Functional Options**: For flexible constructor configuration in Go
