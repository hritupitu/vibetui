.PHONY: build run clean

build:
	go build -o vibetui .

run: build
	./vibetui

clean:
	rm -f vibetui ode
