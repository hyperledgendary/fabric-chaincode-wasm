# Smart Contracts - running in Wasm

Web Assembly (Wasm) is "a new code format for deploying programs that is portable, safe, efficient, and universal.”

This repo provides the PoC for a writing a smart contract runs inside a Wasm engine. This Wasm engine is hosted by a golang chaincode that implements the gRPC to talk back to the peer, and route calls to the guest code in the Wasm engine.

## Usage

This is an early prototype for running Wasm chaincode as an external service using the [Life](https://github.com/perlin-network/life) WebAssembly VM.

It can be used with the [fabric-builders](https://github.com/hyperledgendary/fabric-builders) builder project. The instructions below assume that you have a Fabric network configured to use `hyperledgendary/fabric-builder-peer` images. See the "Chaincode as an external service" documentation for more information.

**Note:** each organization in a Fabric network will need to follow the instructions below to host their own instance of the Wasm chaincode external service.

### Packaging and installing the Wasm chaincode

Create a `connection.json` file with details of how Fabric will connect to your external service. For example:

```
{
  "address": "wasmcc.example.com:9999",
  "dial_timeout": "10s",
  "tls_required": false
}
```

Package the `connection.json` file using the [pkgcc.sh](https://github.com/hyperledgendary/fabric-builders/blob/master/tools/pkgcc.sh) script. For example:

```
pkgcc.sh -l wasmftw -t external connection.json
```

This should produce a `wasmftw.tgz` file which can be installed using the `peer lifecycle chaincode install` command.

### Running the Wasm chaincode external service

To run the service in a container, build a FabCar docker image:

```
docker build -t hyperledgendary/fabric-chaincode-wasm .
```

The Wasm chaincode requires three environment variables to run, `CHAINCODE_SERVER_ADDRESS`, `CHAINCODE_ID`, and `CHAINCODE_WASM_FILE`, which are described in the `chaincode.env.example` file. Copy the example file to `chaincode.env` and edit it before starting the Wasm chaincode container.

Once you have edited the `chaincode.env` file, start the container using the `docker run` command. For example,

```
docker run -it --rm -v ${PWD}/contracts:/local:ro --name wasmcc.example.com --hostname wasmcc.example.com --env-file chaincode.env --network=small_fabricdev hyperledgendary/fabric-chaincode-wasm
```

**Note:** this assumes you are running a Fabric network using `docker-compose`. The value of `--network` will depend on your configuration

### Using the Wasm chaincode

Once you have installed and started the Wasm chaincode, you'll need to approve and commit it as usual. It should then work in exactly the same was as any other chaincode.
