# AGENTS.md

This file provides guidance to AI coding assistants when working with code in this repository.

## Project Overview

The Knative client `kn` is a CLI tool for managing Knative Serving and Eventing resources. It provides a kubectl-plugin-like architecture with full support for Knative Serving (services, revisions, traffic splits) and Knative Eventing (sources, triggers, brokers, channels).

This is a Go-based project (Go 1.24.2+) that uses Cobra for CLI commands and interacts with Kubernetes clusters via client-go.

## Building and Testing

### Build Commands

Use `hack/build.sh` for all build operations:

- `hack/build.sh` - Full build: compile, test, generate docs, format source
- `hack/build.sh -f` - Fast compile only
- `hack/build.sh -f -t` - Compile and test
- `hack/build.sh -t` - Run unit tests only
- `hack/build.sh -c` - Update dependencies, regenerate docs, format source
- `hack/build.sh --lint` - Run linter
- `hack/build.sh -w` - Watch mode for automatic recompilation
- `hack/build.sh -w -t` - Watch mode with tests
- `hack/build.sh -x` - Cross-compile for all platforms
- `hack/build.sh -p GOOS GOARCH` - Cross-compile for specific platform

The binary is output to `./kn` in the current directory.

### Testing

**Unit tests:**
```bash
hack/build.sh -t
```

**E2E tests (requires running Knative cluster with kn in $PATH):**
```bash
test/local-e2e-tests.sh
```

**Run specific E2E test:**
```bash
test/local-e2e-tests.sh -run ^TestBasicWorkflow$
```

**Run only serving or eventing tests:**
```bash
E2E_TAGS="serving" test/local-e2e-tests.sh
E2E_TAGS="eventing" test/local-e2e-tests.sh
```

**Short mode (excludes large-scale tests):**
```bash
test/local-e2e-tests.sh -short
```

Unit tests are alongside the code. E2E tests require `-tags=e2e` and are in `test/e2e/`.

## Code Architecture

### High-Level Structure

The codebase follows a layered architecture:

1. **Entry point** (`cmd/kn/main.go`): Handles plugin discovery, context sharing, and root command execution
2. **Root command** (`pkg/root/`): Assembles all command groups (Serving, Eventing, Other)
3. **Command layer** (`pkg/commands/`): Individual command implementations organized by resource type
4. **Client layer** (`pkg/serving/`, `pkg/eventing/`): Thin abstraction over Kubernetes client-go
5. **Utilities** (`pkg/util/`, `pkg/wait/`, `pkg/flags/`, etc.): Shared functionality

### Command Structure Convention

Commands follow the pattern: `kn <noun> [<noun2>] <verb> [<id>] [--flags]`

- **Nouns** are resource types (e.g., `service`, `revision`, `trigger`, `broker`)
- **Verbs** are CRUD operations: `create`, `update`, `delete`, `describe`, `list`, `apply`
- Commands are organized in command groups defined in `pkg/root/root.go`:
  - **Serving Commands:** service, revision, route, domain, container
  - **Eventing Commands:** source, broker, trigger, channel, subscription, eventtype
  - **Other Commands:** plugin, secret, completion, version

See `conventions/cli.md` for detailed CLI design principles that must be followed.

### Plugin System

Plugins extend `kn` functionality following kubectl's plugin model:

- Plugin manager (`pkg/plugin/manager.go`) discovers plugins in configured directories
- Plugins are executables named `kn-<plugin-name>` in the plugin path or system PATH
- Context sharing feature allows plugins to receive context data from the main `kn` binary
- Plugins cannot override built-in commands (validated at startup)

### Client Abstraction

The codebase uses thin client wrappers around Kubernetes client-go:

- `pkg/serving/v1/client.go`: Interface `KnServingClient` for Serving resources (services, revisions, routes)
- `pkg/eventing/v1/client.go`: Interface `KnEventingClient` for Eventing resources (triggers, brokers)
- Similar patterns in `pkg/sources/`, `pkg/dynamic/`, `pkg/messaging/`

These clients provide:
- CRUD operations on resources
- Retry logic with conflict resolution (`UpdateServiceWithRetry`, `UpdateTriggerWithRetry`)
- Waiting/watching capabilities for async operations
- Three-way merge for `apply` operations (like `kubectl apply`)

### Wait Logic

The `pkg/wait/` package provides utilities for waiting on resource readiness:
- Services wait for ready condition on revisions
- Configurable timeouts and error windows
- Message callbacks for progress updates during waiting

### Flag Handling

Common flag patterns in `pkg/commands/flags/`:
- `sink.go`: Shared sink flag parsing (supports `ksvc:`, `broker:`, `channel:` prefixes)
- `traffic.go`: Traffic splitting flags for services
- `listprint.go`, `listfilters.go`: Output formatting and filtering for list commands
- Binary flags use `--flag`/`--no-flag` pattern (e.g., `--wait`/`--no-wait`)

### Dual Module Structure

The repository uses a dual Go module setup:
- Root module: `knative.dev/client` (CLI tool)
- Nested module: `knative.dev/client/pkg` (library for plugins and external use)
- The root module replaces the pkg module with a local path (`replace knative.dev/client/pkg => ./pkg`)

### Code Generation

Generated code includes:
- Deep copy methods for custom types
- Client code for CRDs
- Documentation (reference manual in `docs/cmd/`)
- Run `hack/build.sh -c` to regenerate

## Important Patterns

### Error Handling

Custom error types in `pkg/errors/`:
- `KNError` interface with `Details()` method for structured error information
- Error wrapping preserves context while providing user-friendly messages

### Service Configuration

`pkg/serving/config_changes.go` defines how service updates are applied:
- Functions like `WithEnv`, `WithImage`, `WithTraffic` modify service specs
- Immutable update pattern: each function returns a new modified service

### Mock Testing

Mock clients are generated for testing:
- `pkg/serving/v1/client_mock.go`
- `pkg/eventing/v1/client_mock.go`
- Tests use these mocks to avoid requiring a real Kubernetes cluster

## Development Notes

- The binary name is extracted dynamically from `os.Args[0]`, allowing the tool to work when renamed
- All commands validate their structure at startup (command groups cannot execute, only leaf commands can)
- The repository includes CLI conventions documentation that should be followed for any new commands
- Configuration is handled via `pkg/config/` with support for config files and environment variables
- The project uses Knative's Serving and Eventing APIs exclusively (works with any Knative installation)
