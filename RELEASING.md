# Release process

Releasing new versions of fabric-chaincode-wasm is currently a manual process.

## Create GitHub release 

Create a new [fabric-chaincode-wasm release](https://github.com/hyperledgendary/fabric-chaincode-wasm/releases) 

## Publish Docker image

Publish a docker image for use when running Wasm chaincode as an external service by following the instructions for [Pushing a Docker container image to Docker Hub](https://docs.docker.com/docker-hub/repos/#pushing-a-docker-container-image-to-docker-hub), e.g.

```
docker build -t hyperledgendary/fabric-chaincode-wasm .
docker tag hyperledgendary/fabric-chaincode-wasm hyperledgendary/fabric-chaincode-wasm:tag
docker push hyperledgendary/fabric-chaincode-wasm
docker push hyperledgendary/fabric-chaincode-wasm:tag
```
