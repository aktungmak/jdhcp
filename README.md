# jdhcp
a dhcp server which provides a callback interface for controlling its behaviour

This package implements a Server, which listens for incoming DHCP messages, parses them and then calls a callback function with the parsed DHCP message.
Based on the output of the callback function, a DHCP message will be sent back to the originating host.

This package does not do any management of addresses, or persistence of other configuration parameters. It is purely a protocol parser and the actual managemnt logic is provided by the user of the package through the callback function.

TODO:
- Add example code showing how this package should be used
- Listen on 255.255.255.255 as well as the provided IP address
- Add parsers for more of the DHCP options
