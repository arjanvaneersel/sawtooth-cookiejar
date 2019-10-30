# Sawtooth Cookie Jar
Simple cookie jar example of a Sawtooth application.

## Introduction
This is a minimal example of a Sawtooth 1.1 application,
with a transaction processor and corresponding client.
This example demonstrates a simple use case, where a baker bakes or eats cookies saved in a virtual cookie jar.

A baker can:
1. bake one or more cookies to add to the cookie jar
2. eat one or more cookies in the cookie jar
3. count the cookies in the cookie jar

All cookie jar transactions have the same 6 hex digit prefix, which is the first 6 hex characters of the SHA-512 hash of "cookiejar" (that is, "a4d219").
The cookie jar is identified by `mycookiejar` with a corresponding public/private keypair.
The cookie jar count is stored at an 70 hex digit address derived from:
* a 6-hex character prefix (the "cookiejar" Transaction Family namespace) and
* the first 64 hex characters of the SHA-512 hash of the "mycookiejar" public key in hex.

## Purpose
The material is made for the introduction to Hyperledger Sawtooth workshop on the 31st of October in Sofia, Bulgaria and is based on the original cookiejar example by Dan Anderson.

## Components
The cookie jar transaction family contains two parts, both having a version in Python 3 and Go:
1. The client application has two parts:
* `pyclient/cookiejar_client.py` or `goclient/client.go`
containing the client class which interfaces to the Sawtooth validator via the REST API
* `pyclient/cookiejar.py` or `goclient/actions.go` and `goclient/main.go` as the Cookie Jar CLI app
The client container is built with files setup.py and respective Dockerfiles.

2. The Transaction Processor, `pyprocessor/cookiejar_tp.py` or `goprocessor/main.go` and `goprocessor/handler.go`

## Docker Usage
### Prerequisites
This example uses docker-compose and Docker containers. If you do not have these installed please follow the instructions here: https://docs.docker.com/install/

**NOTE**

The preferred OS environment is Ubuntu Linux 16.04.3 LTS x64.
Although other Linux distributions which support Docker should work.

### Building Docker containers
To build and run the code
```
sudo docker-compose up --build
```

The `docker-compose.yaml` file creates a genesis block, which contain initial Sawtooth settings, generates Sawtooth and client keys, and starts the Validator, Settings TP, Cookie Jar TP, and REST API.

### Docker client
In a separate shell from above, launch the client shell container:
```
sudo docker exec -it cookiejar-client bash
```
You can locate the correct Docker client container name, if desired, with
`sudo docker ps` .

In the client shell you just started above, run the cookiejar.py application.
Here are some sample commands:
```
cookiejar(.py) bake 100  # Add 100 cookies to the cookie jar
cookiejar(.py) eat 50    # Remove 50 cookies from the cookie jar
cookiejar(.py) count     # Display the number of cookies in the cookie jar
```

To stop the validator and destroy the containers, type `^c` in the docker-compose window, wait for it to stop, then type
```
sudo docker-compose down
```

## Simple Events Handler
A simple events handler is included.  To run, start the validator then
type the following on the command line:
`./events/events_client.py`

A version in Go is also included.

## Exercises for the User
* Add a new function, `empty` which empties the cookie jar (sets the count to 0) in the client and processor
* Add the ability to specify the cookie jar owner key (client only).  Use
[Simplewallet](https://github.com/askmish/sawtooth-simplewallet) as an example
* Replace simple CSV serialization with [CBOR](http://cbor.io/) serialization in the client and processor.
Use the Sawtooth
["inkey"](https://github.com/hyperledger/sawtooth-core/tree/master/sdk/examples/intkey_python)
example application as a pattern.
* Replace simple CSV serialization with [Protobuf](https://developers.google.com/protocol-buffers/) serialization in the client and processor.
Use the Sawtooth
["XO"](https://github.com/hyperledger/sawtooth-core/tree/master/sdk/examples/xo_python)
example application as a pattern
* Translate a transaction processor into another programming language.
See
[Simplewallet](https://github.com/askmish/sawtooth-simplewallet)
and
[Sawtooth SDK examples](https://github.com/hyperledger/sawtooth-core/tree/master/sdk/examples)
* Also translate the Python client into another programming language.
Note that the client and transaction processor do not need to be written in the same language

## License
This example and Hyperledger Sawtooth software are licensed under the [Apache License Version 2.0](LICENSE) software license.

![Photo of sawtooth cookie cutters]( images/sawtooth-cookie-cutters.jpg "Sawtooth cookie cutters")
<br /> *Antique sawtooth cookie cutters.*

© Copyright 2018-2019, Intel Corporation and Arjan van Eersel.
