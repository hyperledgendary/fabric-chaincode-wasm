// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/hyperledgendary/fabric-chaincode-wasm/proxychaincode"
	"github.com/hyperledgendary/fabric-chaincode-wasm/wasmruntime"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/markbates/pkger"
)

// helper function to check for errors
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// starting point of the Wasm Chaincode
//
func main() {
	log.Printf("[host] Wasm Contract runtime..")

	// Step1: We need the Wasm binary bytes, this is using a prepackaged form of the
	// bytes for test purposes
	info, err := pkger.Stat("/contracts/fabric_contract.wasm")
	check(err)

	fmt.Println("Name: ", info.Name())
	fmt.Println("Size: ", info.Size())
	fmt.Println("Mode: ", info.Mode())
	fmt.Println("ModTime: ", info.ModTime())

	f, err := pkger.Open("/contracts/fabric_contract.wasm")
	check(err)
	defer f.Close()

	wasmBytes, err := ioutil.ReadAll(f)
	check(err)

	// start the Wasm runtime, providing the code to execute
	// this will also set up the import/export functions to allow generic
	// marhsalling of operations into and out of Wasm using the Wapc approach
	wrt := wasmruntime.NewRuntime(wasmBytes)

	// create a proxy chaincode to route the invoke/init calls
	cc := proxychaincode.NewChaincode(wrt)

	// start the chaincode shim
	err = shim.Start(cc)
	check(err)

	// everything is now done in response to gRPC calls from the peer
	return
}
