# define vars
BINARY_NAME = proxymgr

# default target
all: build

# build app
build:
	go build -o $(BINARY_NAME) .

# test app
test:
	go test -v ./...

# install app
install:
	cp $(BINARY_NAME) /usr/sbin/
	chown root: /usr/sbin/$(BINARY_NAME)
	chmod 755 /usr/sbin/$(BINARY_NAME)
	cp ./etc/proxymanager.yml /etc/
	chown root: /etc/proxymanager.yml
	chmod 644 /etc/proxymanager.yml