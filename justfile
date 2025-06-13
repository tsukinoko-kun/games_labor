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
