.PHONY: build run

build:
	cd .. && make snapshot
	rm -rf builds/*
	cp ../dist/opsani_*.{tar.gz,zip} builds/
	docker-compose build

run:
	docker-compose up

default: build
