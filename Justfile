default:
    @just --list --justfile {{justfile()}}

tidy:
    go mod tidy

build:
    go build -o ./out/zproxy ./main.go

run:
    go run ./main.go
