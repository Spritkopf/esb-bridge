# esb-bridge


## WORK IN PROGRESS
This is work in progress and not considered functional, as long as this notice exists

## Packages in this module

### internal/usbprotocol
Handles USB serial connection to the esb-bridge-fw (device running the firmware). Should not be used directly

### pkg/esbbridge
Communicates with the firmware to transfer ESB messages to targets, and to receive incoming ESB messages

### cmd/server
CLI tool that Provides an interface to the esb bridge over the network (TCP socket). This is necessary because only one process can access the USB serial port. Also, there is only one physical instance of this device connected to the PC running the server, but there may be several nodes distributed across the network which want to access the ESB devices

### pkg/esbbridgeclient
Talks to the server over TCP socket in order to send and receive ESB messages. This component will be used by end-point implementations, meaning packages that provide access to a class of ESB device (e.g. binary sensor, switch, light etc) or more general packages like a MQTT-to-esb-bridge

## Limitations
* ESB connection parameters are fixed in esb-bridge firmware, cannot be changed

## License
If not mentioned otherwise in individual source files, the MIT license (see LINCESE file) is applicable