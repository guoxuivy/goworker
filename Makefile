.PHONY:build/cxe build/cxe.o
DIR=$(shell pwd)
BUILD=$(DIR)/build
LDFLAGS=-ldflags "-X=main.Version=0.0.1"

build: build/cxe
	@if [ ! -e $(BUILD)/config.yaml ]; then cp $(DIR)/config_bak.yaml $(BUILD)/config.yaml; fi
	go build $(LDFLAGS) -o $(BUILD)/cxe .
	
clean:
	@rm $(BUILD)/cxe*
	
test:
	go test ./...
	
run-area-logs:	
	./build/cxe --config $(BUILD)/config.yaml area --logs | tee $(BUILD)/run-area-logs.log

run-area-daily:	
	./build/cxe --config $(BUILD)/config.yaml area --daily | tee $(BUILD)/run-area-daily.log
	