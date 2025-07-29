.DEFAULT_GOAL := run

BINARY := bin/go-shorty

# Run the application with a build binary(can also just run `go run main.go`)
run: $(BINARY)
	./$(BINARY)

# Build the target BINARY; by making this target the title of the BINARY, make
# will ignore rebuilding the binary if it is available. TODO: implement build
# caching logic?
$(BINARY): main.go
	go build -o $(BINARY) main.go

clean:
ifeq ($(OS),Windows_NT)
	rmdir /S /Q bin
else
	rm -rf bin
endif

lint:
	

test:
	go test -v ./...
