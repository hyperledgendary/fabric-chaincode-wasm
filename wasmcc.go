// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/hyperledgendary/fabric-chaincode-wasm/wasmruntime"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

// ChaincodeConfig is used to configure the chaincode server. See chaincode.env.example
type ChaincodeConfig struct {
	CCID    string
	Address string
	WasmCC  string
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	log.Printf("[host] Wasm Contract runtime..")

	config := ChaincodeConfig{
		CCID:    os.Getenv("CHAINCODE_ID"),
		Address: os.Getenv("CHAINCODE_SERVER_ADDRESS"),
		WasmCC:  os.Getenv("CHAINCODE_WASM_FILE"),
	}

	wasmBytes, err := ioutil.ReadFile(config.WasmCC)
	check(err)

	wrt := wasmruntime.NewRuntime(wasmBytes)

	server := &shim.ChaincodeServer{
		CCID:    config.CCID,
		Address: config.Address,
		CC:      wrt,
		TLSProps: shim.TLSProperties{
			Disabled: true,
		},
	}

	err = server.Start()
	check(err)

	return
}
