# esb-bridge

## WORK IN PROGRESS
This is work in progress and not considered functional, as long as this notice exists

## Todo
[X] Listener thread in ESB module for async RX messages
[ ] ESB Messagebuilder trait: from_usb and try_into, logic needs to go into esb-bridge module, that should implement the trait gluing usb and esb together

## package contents

### esb_bridge --- IN PROGRESS
Library crate containing the logic communicating with the firmware to transfer ESB messages to targets, and to receive 
incoming ESB messages

### usbprotocol
Module that handles USB serial connection to the esb-bridge-fw (device running the firmware). Used by esbbridge crate

### Server -- PLANNED
Binary crate; idk maybe a MQTT bridge?

### App modules -- PLANNED
Modueles for different device classes like sensors, switches and so on

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
