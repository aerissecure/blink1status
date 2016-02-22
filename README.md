# Blink1Status

Blink1Status is a simple program we use with the pentest devices we ship to our
customers for internal testing. It uses a Thingm Blink(1) USB device to help
quickly diagnosis connectivity issues without giving customers credentials to
the device.

## Installation

The easiest way to use the program is to download the static binary from the
[*releases* section](https://github.com/aerissecure/blink1status/releases).

If you'd like to compile it, you may use go get but you must also have the
libusb-dev package to build the go-blink1 library:

    apt-get install libusb-dev
    go get github.com/aerissecure/blink1status

## Usage

Blink1Status comes with 4 simple troubleshooting steps, each indicated by a
different color light:

1. Red: network interface is in "down" state
2. yellow: default gateway is inaccessible with ping
3. purple: both VPN tunnels inaccessible with ping
4. blue: internet is not accessible with ping

Since these states are not mutually exclusive, the Blink1 will continually
cycle through all statuses that are currently in a down state.

Sensible defaults are used except for the VPN tunnel:

- -iface: eth0 is used if network interface is not specified
- -fw: default gateway is detected if one is not specified
- -inet: 8.8.8.8 is used as internet endpoint if one is not specified

Example:

    ./blink1status -tun1 10.0.0.1

