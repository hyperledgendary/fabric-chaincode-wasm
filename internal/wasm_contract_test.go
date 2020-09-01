// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package internal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"

	"github.com/hyperledgendary/fabric-chaincode-wasm/internal"
	"github.com/hyperledgendary/fabric-chaincode-wasm/internal/fakes"
	contract "github.com/hyperledgendary/fabric-ledger-protos-go/contract"
)

var _ = Describe("WasmContract", func() {
	var (
		wasmContract *internal.WasmContract
		wasmInvoker  *fakes.WasmGuestInvoker
	)

	BeforeEach(func() {
		contextStore := internal.NewContextStore()
		wasmInvoker = &fakes.WasmGuestInvoker{}

		wasmContract = internal.NewWasmContract(contextStore, wasmInvoker)
	})

	Describe("Invoke", func() {

		Context("With a successful transaction response", func() {
			var stub *fakes.ChaincodeStubInterface

			BeforeEach(func() {
				stub = &fakes.ChaincodeStubInterface{}

				itr := &contract.InvokeTransactionResponse{}
				itr.Payload = []byte("bond")
				response, _ := proto.Marshal(itr)

				wasmInvoker.InvokeWasmOperationReturns(response, nil)
			})

			It("should return a shim.Success with the correct payload", func() {
				result := wasmContract.Invoke(stub)
				Expect(result.Status).To(Equal(int32(200)))
				Expect(result.Payload).To(Equal([]byte("bond")))
			})
		})
	})
})
