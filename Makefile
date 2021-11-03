sm:
	git submodule update --init language-list

conf: sm

build:
	go run *.go
