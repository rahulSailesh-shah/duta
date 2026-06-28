.PHONY: run dev

run:
	go run .

dev:
	@command -v air >/dev/null 2>&1 || go install github.com/air-verse/air@latest
	air
