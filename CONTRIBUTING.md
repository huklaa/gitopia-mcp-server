# Contributing to Gitopia MCP Server

Thank you for your interest in contributing to the Gitopia MCP Server! This document provides guidelines and information for contributors.

## Getting Started

### Prerequisites

- Go 1.25 or later
- Docker (for containerized development)
- Git
- A Gitopia wallet with mnemonic for testing

### Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/gitopia/gitopia-mcp-server.git
   cd gitopia-mcp-server
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment**
   ```bash
   cp env.example .env
   # Edit .env with your configuration
   ```

4. **Run tests**
   ```bash
   go test ./...
   ```

## Development Workflow

### 1. Issue Creation
- Check existing issues before creating new ones
- Use clear, descriptive titles
- Provide detailed descriptions with context
- Add appropriate labels and milestones

### 2. Branch Naming
- `feature/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation updates
- `refactor/description` - Code refactoring

### 3. Commit Messages
Follow [Conventional Commits](https://www.conventionalcommits.org/):
```
type(scope): description

feat(api): add new repository listing endpoint
fix(auth): resolve wallet authentication issue
docs(readme): update installation instructions
```

### 4. Pull Requests
- Create detailed PR descriptions
- Link related issues
- Include testing instructions
- Ensure CI passes
- Request appropriate reviewers

## Code Standards

### Go Code Style
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Run `golangci-lint` for linting
- Write comprehensive tests

### Testing
- Unit tests for all new functionality
- Integration tests for API endpoints
- Mock external dependencies
- Maintain >80% code coverage

### Documentation
- Update README.md for user-facing changes
- Add inline comments for complex logic
- Update API documentation
- Include examples where helpful

## Project Structure

```
├── cmd/                    # Application entry points
├── internal/              # Internal packages
│   ├── config/           # Configuration management
│   ├── handler/          # MCP tool handlers
│   ├── gitopia/          # Gitopia API integration
│   └── ...
├── docs/                 # Documentation
├── .github/              # GitHub workflows
└── ...
```

## Testing

### Running Tests
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/handler
```

### Integration Tests
```bash
# Set up test environment
export GITOPIA_MNEMONIC="your test mnemonic"
export GITOPIA_GRPC_ENDPOINT="gitopia-grpc.polkachu.com:11390"

# Run integration tests
go test -tags=integration ./...
```

## Building and Deployment

### Local Build
```bash
go build -o bin/gitopia-mcp-server ./cmd/server
```

### Docker Build
```bash
docker build -t gitopia-mcp-server:latest .
```

### Release Process
1. Update CHANGELOG.md
2. Create release PR
3. Tag release after merge
4. GitHub Actions handles Docker image publishing

## Getting Help

- **Issues**: Create GitHub issues for bugs and feature requests
- **Discussions**: Use GitHub Discussions for questions and ideas
- **Documentation**: Check the `docs/` directory for detailed guides

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). Please be respectful and inclusive in all interactions.

## License

By contributing to this project, you agree that your contributions will be licensed under the MIT License.
