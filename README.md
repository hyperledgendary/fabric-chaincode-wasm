# Smart Contracts - running in Wasm

Web Assembly (Wasm) is "a new code format for deploying programs that is portable, safe, efficient, and universal.”

This repo provides the PoC for a writing a smart contract runs inside a Wasm engine. This Wasm engine is hosted by a golang chaincode that implements the gRPC to talk back to the peer, and route calls to the guest code in the Wasm engine.
