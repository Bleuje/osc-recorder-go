### Build

`go build -o osc-recorder`

### Run

example:

`./osc-recorder --address=127.0.0.1 --port=8000 --file=recordings.json --scheme=basic --repeaters=8001,8002 --quantized`

### Options

- --address (required): The address to listen to.
- --port (required): The port to listen to.
- --file (required): Path to the final file that will contain the recording.
- --scheme (required): Specify a specific schema.
- --repeaters: Comma-separated list of ports to forward the received messages to. If not specified, no forwarding will occur.
- --quantized: Quantize the recording so that the first message starts at time 0.
