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

## Schemes

Schemes define how the OSC messages are processed and recorded. You can specify a scheme using the `--scheme` option. The available schemes are:

- `dirt_basic`: A basic schema for recording OSC messages, taking the first element of the args.
- `dirt_strip`: Returns only odd arguments from the args.
- `basic`: A basic schema for recording OSC messages, including all arguments.
- `only_numbers`: Returns only numerical arguments (integers and floats) from the args.

Each scheme is defined as a function that processes the OSC message's address and arguments, returning a dictionary with the processed data. You can extend or modify these schemes by editing the `ALL_SCHEMES` dictionary in the code.


### Dependency

Uses https://github.com/hypebeast/go-osc
