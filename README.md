## osc recorder/replayer in Go

Tool to record osc messages into a json, also has a replayer of recorded jsons.

The recording part is doing what this python repo was doing: https://github.com/Bubobubobubobubo/smol-osc-recorder

### Build

`go build -o osc-recorder`

`go build -o osc-replayer replayer.go`

### Run

example:

`./osc-recorder --address=127.0.0.1 --port=8000 --file=recordings.json --scheme=basic --repeaters=8001,8002 --quantized`

`./osc-replayer --file recordings.json --address 127.0.0.1 --port 8000 --speed 1.0`

### Options

- --address (required): The address to listen to.
- --port (required): The port to listen to.
- --file (required): Path to the final file that will contain the recording.
- --scheme (required): Specify a specific schema.
- --repeaters: Comma-separated list of ports to forward the received messages to. If not specified, no forwarding will occur.
- --quantized: Quantize the recording so that the first message starts at time 0.

### Dependency

Uses https://github.com/hypebeast/go-osc
