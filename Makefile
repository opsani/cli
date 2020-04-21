BIN="./bin"
SRC=$(shell find . -name "*.go")

ifeq (, $(shell which addlicense))
$(warning "could not find addlicense in $(PATH), run: go get -u github.com/google/addlicense")
endif

TEST_RUNNER="richgo"
ifeq (, $(shell which richgo))
$(warning "could not find richgo in $(PATH), run: go get -u github.com/kyoh86/richgo")
TEST_RUNNER="go"
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

.PHONY: test_unit
test_unit:
	$(info ******************** running unit tests ********************)
	$(TEST_RUNNER) test -v ./command/... ./opsani/...

.PHONY: test_integration
test_integration:
	$(info ******************** running integration tests ********************)
	$(TEST_RUNNER) test -v ./integration/...

.PHONY: test
test: test_unit test_integration

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
