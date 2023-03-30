# Gossip
Gossip is a simple Go service, which instances can communicate with each other via a bidirectional gRPC stream by telling each other gossip.

Every told and heard gossip is being logged.

## Configuration
The service configuraion is done via environment variables:
- `HOSTPORT=0.0.0.0:5051` makes the gRPC server listen on the given host and port (the example represents the default value).
- `NEIGHBOURS="192.168.0.10:5051,192.168.0.11:5051"` describes a comma-separated addresses of other services in the network that the current one should interact with.
- `POLLINT=5s` is the duration value determines how frequently the service will try to establish gRPC stream with given neighbours (numbers with `h`, `m`, `s`, `ms` suffixes and their combinations). The default value is `5s` (5 seconds). Service won't start and will log an error in case of having a value with incorrect format.

## Running
### Docker
The simpliest way to see how it works is to run `docker-compose up` command. The command will build and run 5 containers with instances of the service that are configured to interact with each other. The log of sent and received messages along with polling and disconnect messages will be printing to the console.
It's possible to stop/start and pause/unpause individual instances of the servise using the `docker` `stop`, `start`, `pause`, `unpause` followed by the container name corresponding to the instance. You can get the list of the containers' names by using the `docker ps` command. Changes of the services' states should affect the console output.

### Manual run
It's also possible to test how the service works using the terminal. For example, we could open three terminal windows/tabs/splits and run each of the following commands in each of them:
```bash
HOSTPORT="0.0.0.0:5051" NEIGHBOURS="0.0.0.0:5052,0.0.0.0:5053" go run main.go
HOSTPORT="0.0.0.0:5052" NEIGHBOURS="0.0.0.0:5051,0.0.0.0:5053" go run main.go
HOSTPORT="0.0.0.0:5053" NEIGHBOURS="0.0.0.0:5051,0.0.0.0:5052" go run main.go
```
You're welcome to play with the system by using ctrl+C to stop the service and the same command to run it again.