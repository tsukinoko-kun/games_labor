prebuild:
    pnpm run styles
    templ generate
    pnpm run islands

run:
    @just prebuild
    go run ./cmd/app --port 4321

dev:
    @just prebuild
    air

build:
    @just prebuild
    go build -o bin/app ./cmd/app

buildall:
    @just prebuild
    GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o bin/apple_arm64 ./cmd/app
    GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o bin/apple_intel ./cmd/app
    GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o bin/windows.exe ./cmd/app
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/linux_x86 ./cmd/app
    GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bin/linux_arm64 ./cmd/app
    GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -o bin/windows_arm64.exe ./cmd/app
