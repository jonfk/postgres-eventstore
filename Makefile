
.PHONY: clean all

all: bin/event_generator

bin/event_generator: event_generator/main.go
	go build -o bin/event_generator ./event_generator/...

clean:
	rm -r bin
