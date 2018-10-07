all: *.go client/*.go lib/*.go
	go build -race
	cd client
	go build -race
