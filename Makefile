BIN="./bin"
SRC=$(shell find . -name "*.go")

ifeq (, $(shell which addlicense))
$(warning "could not find addlicense in $(PATH), run: go get -u github.com/google/addlicense")
endif

.PHONY: build run fmt vet test deps clean license

default: all

all: fmt vet test build	

build: deps
	$(info ******************** building cli ********************)
	go build -o $(BIN)/opsani main.go

run:
	go run main.go
	
fmt:
	$(info ******************** checking formatting ********************)
	@test -z $(shell gofmt -l $(SRC)) || (gofmt -d $(SRC); exit 1)

test: deps vet
	$(info ******************** running tests ********************)
	ginkgo -r -skipPackage vendor --randomizeAllSpecs --randomizeSuites --failOnPending

deps:
	$(info ******************** downloading dependencies ********************)
	go get -v ./...

vet:
	$(info ******************** vetting ********************)
	go vet ./...

clean:
	rm -rf $(BIN)

install: build
	$(info ******************** installing ********************)
	cp $(BIN)/opsani /usr/local/bin/opsani

completion:
	$(info ******************** completion ********************)
	go run . completion --shell zsh > /usr/local/share/zsh-completions/_opsani

license:
	$(info ******************** licensing ********************)
	addlicense -c "Opsani" -l apache -v Dockerfile *.go ./**/*.go
