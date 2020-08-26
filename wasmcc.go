// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hyperledgendary/fabric-chaincode-wasm/internal"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

// ChaincodeConfig is used to configure the chaincode server. See chaincode.env.example
type ChaincodeConfig struct {
	CCID    string
	Address string
	WasmCC  string
}

func main() {
	log.Printf("[host] Wasm Chaincode client-server...\n")

	config := ChaincodeConfig{
		CCID:    os.Getenv("CHAINCODE_ID"),
		Address: os.Getenv("CHAINCODE_SERVER_ADDRESS"),
		WasmCC:  os.Getenv("CHAINCODE_WASM_FILE"),
	}
	log.Printf("[host] CCID: %s\n", config.CCID)
	log.Printf("[host] Address: %s\n", config.Address)
	log.Printf("[host] WasmCC: %s\n", config.WasmCC)

	contextStore := internal.NewContextStore()
	proxy := internal.NewFabricProxy(contextStore)

	wasmGuest, err := internal.NewWasmGuest(config.WasmCC, proxy)
	if err != nil {
		panic(err)
	}
	defer wasmGuest.Close()

	contract := internal.NewWasmContract(contextStore, wasmGuest)

	if len(config.Address) > 0 {
		log.Printf("[host] Wasm Chaincode server starting...\n")
		server := &shim.ChaincodeServer{
			CCID:    config.CCID,
			Address: config.Address,
			CC:      contract,
			TLSProps: shim.TLSProperties{
				Disabled: true,
			},
		}

		if err := server.Start(); err != nil {
			fmt.Printf("Error starting Wasm chaincode server: %s", err.Error())
		}
	} else {
		log.Printf("[host] Wasm Chaincode starting...\n")
		if err := shim.Start(contract); err != nil {
			fmt.Printf("Error starting Wasm chaincode: %s", err.Error())
		}
	}

	log.Printf("[host] Wasm Chaincode done\n")
}
