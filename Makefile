run: build
	./app

build:
	go build -o app

create-file:
	dd if=/dev/zero of=file.txt count=1024 bs=1048576