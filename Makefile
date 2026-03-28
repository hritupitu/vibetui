.PHONY: build run clean

build:
	go build -o ode ./...

run: build
	./ode

clean:
	rm -f ode
