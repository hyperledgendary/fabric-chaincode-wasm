# Copyright the Hyperledger Fabric contributors. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

ARG GO_VER=1.13.8

# Alpine image doesn't work for wasmer :(
FROM golang:${GO_VER}

WORKDIR /go/src/github.com/hyperledgendary/fabric-chaincode-wasm
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

EXPOSE 9999
CMD ["fabric-chaincode-wasm"]
