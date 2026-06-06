BIN_DIR := bin
CMD_DIR := cmd

build:
	@echo "Building binaries..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/olivine ./$(CMD_DIR)/olivine

clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
