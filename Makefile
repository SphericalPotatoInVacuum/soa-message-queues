compose_path := ./deploy/docker-compose.yml
env_path := ./.env
compose := docker-compose -f "$(compose_path)"


help: # Print help on Makefile
	@grep '^[^.#]\+:\s\+.*#' Makefile | \
	sed "s/\(.\+\):\s*\(.*\) #\s*\(.*\)/`printf "\033[93m"`\1`printf "\033[0m"`	\3 [\2]/" | \
	expand -t20

proto_gen: $(shell find proto -type f -name "*.proto")
	@echo " ====== Building protobufs ======"
	@sh -c "$(CURDIR)/scripts/generate_pb.sh"

protobuf: proto_gen # Build protobufs

build: protobuf # Build containers
	@echo " ====== Building containers ======"
	@$(compose) --profile client build

push: build # Push containers to the registry
	@echo " ====== Pushing images to the registry ======"
	@$(compose) --profile client push

up: # Start server
	@echo " ====== Launching server ======"
	-@$(compose) up --scale grabber=8 rabbitmq grabber server
	@$(MAKE) clean

test: build # Test application
	@echo " ====== Running tests ======"
	@$(compose) up -d rabbitmq grabber server
	-@$(compose) run --rm -it client
	@$(MAKE) clean

clean: # Clean everything
	@echo " ====== Tearing down ======"
	@$(compose) -v down

.PHONY: help protobuf build up test clean
