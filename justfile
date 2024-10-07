_run_tests:
    go test -coverprofile=coverage.out ./...

_cleancoverage:
    rm coverage.out

test: _run_tests _cleancoverage
cover: _run_tests
    go tool cover -html=coverage.out
    just _cleancoverage


build:
    ./build.sh
