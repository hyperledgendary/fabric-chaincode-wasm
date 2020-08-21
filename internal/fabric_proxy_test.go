// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package internal_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/hyperledgendary/fabric-chaincode-wasm/internal"
	"github.com/hyperledgendary/fabric-chaincode-wasm/internal/fakes"
	contract "github.com/hyperledgendary/fabric-ledger-protos-go/contract"
	"google.golang.org/protobuf/proto"
)

var _ = Describe("FabricProxy", func() {
	var (
		contextStore *internal.ContextStore
		proxy        *internal.FabricProxy
		ctx          context.Context
	)

	BeforeEach(func() {
		contextStore = internal.NewContextStore()
		proxy = internal.NewFabricProxy(contextStore)
		ctx = context.Background()
	})

	Describe("FabricCall", func() {

		Context("With an invalid request", func() {
			It("should error if the binding is not wapc", func() {
				result, err := proxy.FabricCall(ctx, "notWapc", "LedgerService", "CreateState", []byte(""))
				Expect(result).To(BeNil())
				Expect(err).To(MatchError("Operation not supported: notWapc LedgerService CreateState"))
			})

			It("should error if the namespace is not LedgerService", func() {
				result, err := proxy.FabricCall(ctx, "wapc", "TeaService", "CreateState", []byte(""))
				Expect(result).To(BeNil())
				Expect(err).To(MatchError("Operation not supported: wapc TeaService CreateState"))
			})

			It("should error with an unsupported operation", func() {
				result, err := proxy.FabricCall(ctx, "wapc", "LedgerService", "MutateState", []byte(""))
				Expect(result).To(BeNil())
				Expect(err).To(MatchError("Operation not supported: wapc LedgerService MutateState"))
			})
		})

		Context("With a CreateState request", func() {
			var payload []byte

			BeforeEach(func() {
				context := &contract.TransactionContext{}
				context.ChannelId = "channel1"
				context.TransactionId = "txn1"
				state := &contract.State{}
				state.Key = "007"
				state.Value = []byte("bond")
				request := &contract.CreateStateRequest{}
				request.Context = context
				request.State = state
				payload, _ = proto.Marshal(request)
			})

			It("should handle missing context error", func() {
				result, err := proxy.FabricCall(ctx, "wapc", "LedgerService", "CreateState", payload)
				Expect(result).To(BeNil())
				Expect(err).To(MatchError("CreateState failed: No stub found for transaction context channel1 txn1"))
			})

			It("should create a new state if it does not already exist", func() {
				stub := &fakes.ChaincodeStubInterface{}
				contextStore.Put("channel1", "txn1", stub)

				Expect(proxy.FabricCall(ctx, "wapc", "LedgerService", "CreateState", payload)).To(BeNil())

				Expect(stub.GetStateCallCount()).To(Equal(1), "Should call GetState once")
				key := stub.GetStateArgsForCall(0)
				Expect(key).To(Equal("007"), "Should call GetState with correct key")

				Expect(stub.PutStateCallCount()).To(Equal(1), "Should call PutState once")
				key, value := stub.PutStateArgsForCall(0)
				Expect(key).To(Equal("007"), "Should call PutState with correct key")
				Expect(value).To(Equal([]byte("bond")), "Should call PutState with correct value")
			})

			It("should fail if the state already exists", func() {
				stub := &fakes.ChaincodeStubInterface{}
				stub.GetStateReturns([]byte("dr evil"), nil)
				contextStore.Put("channel1", "txn1", stub)

				result, err := proxy.FabricCall(ctx, "wapc", "LedgerService", "CreateState", payload)
				Expect(result).To(BeNil())
				Expect(err).To(MatchError("CreateState failed: State already exists for key 007"))

				Expect(stub.GetStateCallCount()).To(Equal(1), "Should call GetState once")
				key := stub.GetStateArgsForCall(0)
				Expect(key).To(Equal("007"), "Should call GetState with correct key")

				Expect(stub.PutStateCallCount()).To(Equal(0), "Should not call PutState")
			})
		})

		Context("With an UpdateState request", func() {
			var payload []byte

			BeforeEach(func() {
				context := &contract.TransactionContext{}
				context.ChannelId = "channel1"
				context.TransactionId = "txn1"
				state := &contract.State{}
				state.Key = "007"
				state.Value = []byte("bond")
				request := &contract.UpdateStateRequest{}
				request.Context = context
				request.State = state
				payload, _ = proto.Marshal(request)
			})

			It("should handle missing context error", func() {
				result, err := proxy.FabricCall(ctx, "wapc", "LedgerService", "UpdateState", payload)
				Expect(result).To(BeNil())
				Expect(err).To(MatchError("UpdateState failed: No stub found for transaction context channel1 txn1"))
			})

			It("should fail if the state does not already exist", func() {
				stub := &fakes.ChaincodeStubInterface{}
				contextStore.Put("channel1", "txn1", stub)

				result, err := proxy.FabricCall(ctx, "wapc", "LedgerService", "UpdateState", payload)
				Expect(result).To(BeNil())
				Expect(err).To(MatchError("UpdateState failed: No state exists for key 007"))

				Expect(stub.GetStateCallCount()).To(Equal(1), "Should call GetState once")
				key := stub.GetStateArgsForCall(0)
				Expect(key).To(Equal("007"), "Should call GetState with correct key")

				Expect(stub.PutStateCallCount()).To(Equal(0), "Should not call PutState")
			})

			It("should update a state which already exists", func() {
				stub := &fakes.ChaincodeStubInterface{}
				stub.GetStateReturns([]byte("dr evil"), nil)
				contextStore.Put("channel1", "txn1", stub)

				Expect(proxy.FabricCall(ctx, "wapc", "LedgerService", "UpdateState", payload)).To(BeNil())

				Expect(stub.GetStateCallCount()).To(Equal(1), "Should call GetState once")
				key := stub.GetStateArgsForCall(0)
				Expect(key).To(Equal("007"), "Should call GetState with correct key")

				Expect(stub.PutStateCallCount()).To(Equal(1), "Should call PutState once")
				key, value := stub.PutStateArgsForCall(0)
				Expect(key).To(Equal("007"), "Should call PutState with correct key")
				Expect(value).To(Equal([]byte("bond")), "Should call PutState with correct value")
			})
		})
	})

})
