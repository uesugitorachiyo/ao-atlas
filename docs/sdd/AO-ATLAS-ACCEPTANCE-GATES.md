# AO Atlas Acceptance Gates

Production readiness requires:

- `go test ./...`
- `go vet ./...`
- `go build ./cmd/atlas`
- JSON syntax validation for all schemas and fixtures
- CLI validation for all valid fixtures
- CLI rejection for invalid fixtures
- public-safety scan over tracked and pending files
- `git diff --check`

The readiness script prints `score=100/100` only after every gate passes.

