# esb-bridge

## WORK IN PROGRESS
This is work in progress and not considered functional, as long as this notice exists

## package contents

### esb_bridge --- IN PROGRESS
Library crate containing the logic communicating with the firmware to transfer ESB messages to targets, and to receive 
incoming ESB messages

### usbprotocol
Module that handles USB serial connection to the esb-bridge-fw (device running the firmware). Used by esbbridge crate

### Server -- IN PROGRESS
Binary crate; CLI tool that provides an interface to the esb bridge over the network (TCP socket). 
This is necessary because only one process can access the USB serial port. Also, there is only one physical instance of 
this device connected to the PC running the server, but there may be several nodes distributed across the network which 
want to access the ESB devices.

### Client -- PLANNED
Module that talks to the server over TCP socket in order to send and receive ESB messages. 
This component can be used by end-point implementations, meaning packages that provide access to a class of 
ESB device (e.g. binary sensor, switch, light etc) or more general packages like a MQTT-to-esb-bridge
May be extended in the future to be a binary create to provide a CLI tool...

## Get it running -- PLANNED
The server is the only component of this repository which is intended to run directly
Examples below: ESB bridge device on /dev/ttyACM0, server on port 9815

### Run executable of the server
```
$ cargo run --bin server -- --device /dev/ttyACM0 --port 9815
```

### Docker -- PLANNED
Build the Docker image
```
$ docker build . -t esb-bridge-server
```

Run the server
```
$ docker run -d --device /dev/ttyACM0 -p 9815:9815 --hostname esbbridgeserver esb-bridge-server
```
Note: The parameter `--hostname esbbridgeserver` is important, without it clients cannot connect to the server. This will be be resolved in a future version (hopefully)


## Limitations
* ESB connection parameters are fixed in esb-bridge firmware, cannot be changed

## License
If not mentioned otherwise in individual source files, the MIT license (see LICENSE file) is applicable
