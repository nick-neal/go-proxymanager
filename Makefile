# define vars
BINARY_NAME = proxymanager

# default target
all: build

# build app
build:
	go build -o $(BINARY_NAME) .