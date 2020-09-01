// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package internal_test

import (
	"context"

	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
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

		Context("When something panics", func() {
			It("should recover and return an error", func() {
				result, err := proxy.FabricCall(ctx, "wapc", "LedgerService", "CreateState", nil)
				Expect(result).To(BeNil())
				Expect(err).To(MatchError("Operation panicked: wapc LedgerService CreateState"))
			})
		})

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

		Context("With a GetStatesRequest_ByKeyRange request", func() {
			var (
				request *contract.GetStatesRequest
				query   *contract.GetStatesRequest_ByKeyRange
			)

			BeforeEach(func() {
				context := &contract.TransactionContext{}
				context.ChannelId = "channel1"
				context.TransactionId = "txn1"
				query = &contract.GetStatesRequest_ByKeyRange{}
				request = &contract.GetStatesRequest{}
				request.Context = context
			})

			It("should handle missing context error", func() {
				keyRangeQuery := &contract.KeyRangeQuery{}
				keyRangeQuery.StartKey = "001"
				keyRangeQuery.EndKey = "009"
				query.ByKeyRange = keyRangeQuery
				request.Query = query
				payload, _ := proto.Marshal(request)

				result, err := proxy.FabricCall(ctx, "wapc", "LedgerService", "GetStates", payload)
				Expect(result).To(BeNil())
				Expect(err).To(MatchError("GetStates failed: No stub found for transaction context channel1 txn1"))
			})

			It("should get the specified range of states", func() {
				sqi := &fakes.StateQueryIteratorInterface{}
				sqi.HasNextReturnsOnCall(0, true)
				sqi.HasNextReturnsOnCall(1, true)
				sqi.HasNextReturnsOnCall(2, false)
				sqi.NextReturnsOnCall(0, &queryresult.KV{
					Key:   "007",
					Value: []byte("bond"),
				}, nil)
				sqi.NextReturnsOnCall(1, &queryresult.KV{
					Key:   "008",
					Value: []byte("not bond"),
				}, nil)

				stub := &fakes.ChaincodeStubInterface{}
				stub.GetStateByRangeReturns(sqi, nil)
				contextStore.Put("channel1", "txn1", stub)

				keyRangeQuery := &contract.KeyRangeQuery{}
				keyRangeQuery.StartKey = "001"
				keyRangeQuery.EndKey = "009"
				query.ByKeyRange = keyRangeQuery
				request.Query = query
				payload, _ := proto.Marshal(request)

				result, _ := proxy.FabricCall(ctx, "wapc", "LedgerService", "GetStates", payload)

				Expect(stub.GetStateByRangeCallCount()).To(Equal(1), "Should call GetStateByRange once")
				startKey, endKey := stub.GetStateByRangeArgsForCall(0)
				Expect(startKey).To(Equal("001"), "Should call GetStateByRange with specified start key")
				Expect(endKey).To(Equal("009"), "Should call GetStateByRange with specified end key")

				response := &contract.GetStatesResponse{}
				_ = proto.Unmarshal(result, response)
				states := response.GetStates()
				Expect(len(states)).To(Equal(2))
				Expect(states[0].Key).To(Equal("007"))
				Expect(states[0].Value).To(Equal([]byte("bond")))
				Expect(states[1].Key).To(Equal("008"))
				Expect(states[1].Value).To(Equal([]byte("not bond")))
			})

			It("should handle an unbounded start key", func() {
				sqi := &fakes.StateQueryIteratorInterface{}
				stub := &fakes.ChaincodeStubInterface{}
				stub.GetStateByRangeReturns(sqi, nil)
				contextStore.Put("channel1", "txn1", stub)

				keyRangeQuery := &contract.KeyRangeQuery{}
				keyRangeQuery.EndKey = "009"
				query.ByKeyRange = keyRangeQuery
				request.Query = query
				payload, _ := proto.Marshal(request)

				Expect(proxy.FabricCall(ctx, "wapc", "LedgerService", "GetStates", payload)).NotTo(BeNil())

				Expect(stub.GetStateByRangeCallCount()).To(Equal(1), "Should call GetStateByRange once")
				startKey, endKey := stub.GetStateByRangeArgsForCall(0)
				Expect(startKey).To(Equal(""), "Should call GetStateByRange with an unspecified start key")
				Expect(endKey).To(Equal("009"), "Should call GetStateByRange with specified end key")
			})

			It("should handle an unbounded end key", func() {
				sqi := &fakes.StateQueryIteratorInterface{}
				stub := &fakes.ChaincodeStubInterface{}
				stub.GetStateByRangeReturns(sqi, nil)
				contextStore.Put("channel1", "txn1", stub)

				keyRangeQuery := &contract.KeyRangeQuery{}
				keyRangeQuery.StartKey = "001"
				query.ByKeyRange = keyRangeQuery
				request.Query = query
				payload, _ := proto.Marshal(request)

				Expect(proxy.FabricCall(ctx, "wapc", "LedgerService", "GetStates", payload)).NotTo(BeNil())

				Expect(stub.GetStateByRangeCallCount()).To(Equal(1), "Should call GetStateByRange once")
				startKey, endKey := stub.GetStateByRangeArgsForCall(0)
				Expect(startKey).To(Equal("001"), "Should call GetStateByRange with specified start key")
				Expect(endKey).To(Equal(""), "Should call GetStateByRange with an unspecified end key")
			})
		})
	})

})
