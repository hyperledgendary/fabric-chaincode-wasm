// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/hyperledgendary/fabric-chaincode-wasm/internal"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	wapc "github.com/wapc/wapc-go"
)

// ChaincodeConfig is used to configure the chaincode server. See chaincode.env.example
type ChaincodeConfig struct {
	CCID    string
	Address string
	WasmCC  string
}

func consoleLog(msg string) {
	fmt.Println(msg)
}

func newChaincodePool(wasmFile string, proxy *internal.FabricProxy) (*wapc.Pool, error) {
	wasmBytes, err := ioutil.ReadFile(wasmFile)
	if err != nil {
		return nil, err
	}

	module, err := wapc.New(consoleLog, wasmBytes, proxy.FabricCall)
	if err != nil {
		return nil, err
	}
	// TODO when should this be closed?
	// defer module.Close()

	wapcPool, err := wapc.NewPool(module, 10)
	if err != nil {
		return nil, err
	}
	// TODO when should this be closed?
	// defer wapcPool.Close()

	return wapcPool, nil
}

func main() {
	log.Printf("[host] Wasm Chaincode server...")

	config := ChaincodeConfig{
		CCID:    os.Getenv("CHAINCODE_ID"),
		Address: os.Getenv("CHAINCODE_SERVER_ADDRESS"),
		WasmCC:  os.Getenv("CHAINCODE_WASM_FILE"),
	}

	contextStore := internal.NewContextStore()
	proxy := internal.NewFabricProxy(contextStore)

	wapcPool, err := newChaincodePool(config.WasmCC, proxy)
	if err != nil {
		panic(err)
	}

	contract := internal.NewWasmContract(contextStore, wapcPool)

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
}
