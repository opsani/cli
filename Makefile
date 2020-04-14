BIN="./bin"
SRC=$(shell find . -name "*.go")

ifeq (, $(shell which richgo))
$(warning "could not find richgo in $(PATH), run: go get github.com/kyoh86/richgo")
endif

.PHONY: build run fmt vet test install_deps clean

default: all

all: fmt vet test build	

build: install_deps
	$(info ******************** building cli ********************)
	go build -o $(BIN)/opsani main.go

run:
	go run main.go
	
fmt:
	$(info ******************** checking formatting ********************)
	@test -z $(shell gofmt -l $(SRC)) || (gofmt -d $(SRC); exit 1)

test: install_deps vet
	$(info ******************** running tests ********************)
	richgo test -v ./...

# opsani: install_deps
# 	$(info ******************** building cli ********************)
# 	mkdir -p $(BIN)
# 	make -C cobra all

install_deps:
	$(info ******************** downloading dependencies ********************)
	go get -v ./...

vet:
	$(info ******************** vetting ********************)
	go vet ./...

clean:
	rm -rf $(BIN)
