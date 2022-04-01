# Pathfinder

Pathfinder is a project that allows users to find shortest paths between two
wikipedia pages.

# Overview

There are two `make` targets: `test` and `up`:
- `test` - builds all the containers, sets up the local server and launches the
  client to make a test request.
- `up` - sets up the server workers for deployment

Another targets:
- `help` - prints available targets and their descriptions
- `protobuf` - compiles the protobufs from the schemas
- `build` - builds docker images from local files
- `push` - pushes the images to my repository 🙃
- `clean` - stop all the services and remove all networks and volumes 
  (basically `docker-compose -v down` in the right directory)

# Prerequisites

- Docker
- Docker-compose

# Configuration

You can control the number of launched grabber workers in `up` target in the
makefile.

# Simple start 

```bash
RABBITMQ_USER=guest RABBITMQ_PASS=quest source <(curl -s https://raw.githubusercontent.com/SphericalPotatoInVacuum/soa-message-queues/main/scripts/start_server.sh)
```

# Simple tests

```bash
RABBITMQ_USER=guest RABBITMQ_PASS=quest source <(curl -s https://raw.githubusercontent.com/SphericalPotatoInVacuum/soa-message-queues/main/scripts/run_tests.sh)
```
