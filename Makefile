BINARY   = utility
CMD      = ./cmd/app
BUILD_DIR = ./build

.PHONY: build install clean

build:
	go build -o $(BUILD_DIR)/$(BINARY) $(CMD)

install:
	go install $(CMD)

clean:
	rm -rf $(BUILD_DIR)
