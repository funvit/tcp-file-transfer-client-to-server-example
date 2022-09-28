
.PHONY:build
build: clean
	go build -o ./bin/server ./cmd/server/...
	go build -o ./bin/client ./cmd/client/...

.PHONY:clean
clean:
	rm -fR ./bin/*

.PHONY:fmt
format:
	go fmt ./...
