run: build
	./app

build:
	go build -o app

create-file:
	dd if=/dev/zero of=file2GB.txt count=1024 bs=2097152
