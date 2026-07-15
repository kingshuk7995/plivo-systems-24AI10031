all: sender receiver

sender: cmd/sender/main.go
	go build -o sender ./cmd/sender

receiver: cmd/receiver/main.go
	go build -o receiver ./cmd/receiver

clean:
	rm -rf __pycache__
	rm -f sender receiver
