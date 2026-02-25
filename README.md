# apiconnector

API Connectivity Tester - validates HTTP endpoints and port connectivity for services.

## Purpose

Test connectivity to API endpoints and verify that services are reachable on their configured ports.

## Installation

```bash
go build -o apiconnector ./cmd/apiconnector
```

## Usage

```bash
apiconnector <service1> <service2> ...
```

Format: `name=http://url[:port]`

### Examples

```bash
# Test single endpoint
apiconnector api=http://localhost:8080/health

# Test multiple services
apiconnector api=http://localhost:8080/health db=postgres://localhost:5432

# Test with custom port
apiconnector service=http://example.com:9000/api
```

## Output

```
=== API CONNECTIVITY TEST ===

api                    OK (15ms)
db                     OK (3ms)

Summary: 2 OK, 0 FAIL
```

## Dependencies

- Go 1.21+
- github.com/fatih/color

## Build and Run

```bash
# Build
go build -o apiconnector ./cmd/apiconnector

# Run with dependencies
go run ./cmd/apiconnector api=http://localhost:8080/health
```

## License

MIT