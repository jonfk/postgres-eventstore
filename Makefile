
.PHONY: clean all

all: bin/event_generator bin/event_projector

bin/event_generator: event_generator/main.go
	go build -o bin/event_generator ./event_generator/...

bin/event_projector: event_projector/main.go
	go build -o bin/event_projector ./event_projector/...

clean:
	rm -r bin
