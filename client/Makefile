# It assumes that directory /srv/ is already exists

.PHONY: all deps clean remove-go install uninstall

all: build

deps:
	@echo '# Install Golang'
	apt install golang-go -y

	# Path to go
	echo $(which go)

remove-go:
	@echo '# Uninstall Golang'

	# Remove Go-Build cache
	if [ -d /srv/client_hog ] ; then \
		rm -rf $(HOME)/.cache/go-build; \
	fi

	# Remove Go Path
	if [ -d $(HOME)/go ] ; then \
		rm -rf $(HOME)/go; \
	fi

	# Switched to apt.
	apt purge golang* -y

build: serverBuild

serverBuild:
	@go mod tidy
	@echo '# Build Golang project'
	@go build -o hogBuild .

clean:
	@echo '# Remove compiled server in Github repository'
	if [ -f hogBuild ] ; then \
    		rm hogBuild; \
	fi

install: build
	if [ -d /srv/client_hog ] ; then \
		rm -rf /srv/client_hog; \
	fi

	mkdir /srv/client_hog

	@echo '# Move client_hog.service to systemd'
	cp client_hog.service /etc/systemd/system/

	@echo '# Move Golang Project to /srv/client_hog'
	cp -r hogBuild /srv/client_hog/hog

	@echo '# Start Service'
	systemctl enable --now client_hog.service

uninstall: /srv/client_hog
	@echo '# Remove any existing service installations'
	if [ -h /etc/systemd/system/multi-user.target.wants/client_hog.service ] ; then \
		systemctl stop client_hog; \
		systemctl disable client_hog; \
	fi

	if [ -f /etc/systemd/system/client_hog.service ] ; then \
		rm /etc/systemd/system/client_hog.service; \
	fi

	if [ -d /srv/client_hog ] ; then \
		rm -rf /srv/client_hog/; \
	fi

ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
