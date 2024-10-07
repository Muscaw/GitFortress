_run_tests:
    go test -race -coverprofile=coverage.txt -covermode=atomic ./...

_cleancoverage:
    rm coverage.txt

ci: build _run_tests
test: _run_tests _cleancoverage
cover: _run_tests
    go tool cover -html=coverage.txt
    just _cleancoverage


build:
    ./build.sh
