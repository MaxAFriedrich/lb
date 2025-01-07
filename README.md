# lb

This is a simple web server proxy that unifies a load of backends under a single domain as per the backend map file.

## Building

To build the server, you will need to have go installed. You should be able just clone the repo and then run
`go build .` to build the server.

## Usage

The binary is called `lb` and requires a single argument which is the path to the backend map file.


