# Gossip
Gossip is a simple Go service, which instances can communicate with each other via a bidirectional gRPC streams by telling each other gossip.

Every told and heard gossip is being logged.

## Service discovery
1. After start the service will automatiacally determine available local IP address and broadcast address of the network that it's connected to. 
*Implementation assumes that the host system is only connected to one IP network. In case there are several IP networks available, it will use the random one.*
2. It will send broadcast message using UDP to the network on port 8831 to inform other services on the network of it's presence. For instance, the message can be `gossip://192.168.1.9:5051`. From such messages neighbours should be able to understand that a fellow service is available for communication and how to connect to it.
3. Service will also listen the network on UPD port 8831. Once it receives the message from the neighbour, it will try to connect using gRPC and communicate gossip.  

## Configuration
The service configuraion (optional) is done via environment variables:
- `PORT=5051` makes the gRPC server listen on the given host and port (the default value is `5051`)
- `POLLINT=5s` is the duration value determines how frequently the service will try to establish gRPC stream with given neighbours (numbers with `h`, `m`, `s`, `ms` suffixes and their combinations). The default value is `5s` (5 seconds). Service won't start and will log an error in case of having a value with incorrect format.

## Running
### Docker
The simpliest way to see how it works is to run `docker-compose up` command. The command will build and run 5 replicas of the container with the service that will be interacting with each other. The log of sent and received messages along with discovery and disconnect messages will be printing to the console.
It's possible to stop/start and pause/unpause individual instances of the servise using the `docker` `stop`, `start`, `pause`, `unpause` followed by the container name corresponding to the instance. You can get the list of the containers' names by using the `docker ps` command. Changes of the services' states should affect the console output.