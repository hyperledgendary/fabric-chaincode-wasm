// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"wcr/wasmruntime"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/markbates/pkger"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	log.Printf("[host] Wasm Contract runtime..")

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

	wrt := wasmruntime.NewRuntime(wasmBytes)

	err = shim.Start(wrt)
	check(err)

	return
}
