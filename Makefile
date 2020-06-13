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

ifeq (, $(shell which goreleaser))
$(warning "could not find goreleaser in $(PATH), run: curl -sfL https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh | sh")
endif

ifeq (, $(shell which pkger))
$(warning "could not find pkger in $(PATH), run: go get github.com/markbates/pkger/cmd/pkger")
endif

.PHONY: build run fmt vet test deps clean license snapshot test_integration test_unit image install

default: all

all: fmt vet test build	

build: deps
	$(info ******************** building cli ********************)
	go build -o $(BIN)/opsani main.go

image: 
	$(info ******************** building Docker image ********************)
	docker build . -t opsani/cli:latest

run:
	go run main.go
	
fmt:
	$(info ******************** checking formatting ********************)
	@test -z $(shell gofmt -l $(SRC)) || (gofmt -d $(SRC); exit 1)

test_unit:
	$(info ******************** running unit tests ********************)
	$(TEST_RUNNER) test -v ./command/... ./opsani/...

test_integration:
	$(info ******************** running integration tests ********************)
	$(TEST_RUNNER) test -v ./integration/...

test: test_unit test_integration

deps:
	$(info ******************** downloading dependencies ********************)
	go get -v ./...
	pkger

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

snapshot:
	$(info ******************** snapshotting ********************)
	goreleaser --snapshot --skip-publish --rm-dist

release:
	$(info ******************** releasing ********************)
	goreleaser --rm-dist
	