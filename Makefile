tracker:
	go run ./cmd/tracker

test:
	go test ./... -race -count=1

cover:
	go test ./... -coverprofile=cover.out && go tool cover -html=cover.out

.PHONY: tracker test cover
