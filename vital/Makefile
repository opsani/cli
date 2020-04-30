.DEFAULT_GOAL := default
.PHONY: snapshot app run

snapshot:
	cd .. && make snapshot
	rm -rf assets/builds/*
	cp ../dist/opsani_*.{tar.gz,zip} assets/builds/

app:
	go build .
	docker build -t vital_demo:latest .

run:	
	docker run --rm -it -p 5678:8080 -v `pwd`/.config.yaml:/app/.config.yaml -v `pwd`/credentials.json:/app/credentials.json -v `pwd`/token.json:/app/token.json -v `pwd`/assets:/app/assets vital_demo:latest

build: snapshot app
default: app run
