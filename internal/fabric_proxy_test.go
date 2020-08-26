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

		Context("With a world state CreateState request", func() {
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

			It("should create a new state if it does not already exist in the world state", func() {
				stub := &fakes.ChaincodeStubInterface{}
				contextStore.Put("channel1", "txn1", stub)

				Expect(proxy.FabricCall(ctx, "wapc", "LedgerService", "CreateState", payload)).To(BeNil())

				Expect(stub.GetStateCallCount()).To(Equal(1), "Should call GetState once")
				Expect(stub.GetPrivateDataCallCount()).To(Equal(0), "Should not call GetPrivateData")
				key := stub.GetStateArgsForCall(0)
				Expect(key).To(Equal("007"), "Should call GetState with correct key")

				Expect(stub.PutStateCallCount()).To(Equal(1), "Should call PutState once")
				Expect(stub.PutPrivateDataCallCount()).To(Equal(0), "Should not call PutPrivateData")
				key, value := stub.PutStateArgsForCall(0)
				Expect(key).To(Equal("007"), "Should call PutState with correct key")
				Expect(value).To(Equal([]byte("bond")), "Should call PutState with correct value")
			})

			It("should fail if the state already exists in the world state", func() {
				stub := &fakes.ChaincodeStubInterface{}
				stub.GetStateReturns([]byte("dr evil"), nil)
				contextStore.Put("channel1", "txn1", stub)

				result, err := proxy.FabricCall(ctx, "wapc", "LedgerService", "CreateState", payload)
				Expect(result).To(BeNil())
				Expect(err).To(MatchError("CreateState failed: State already exists for key 007"))

				Expect(stub.GetStateCallCount()).To(Equal(1), "Should call GetState once")
				Expect(stub.GetPrivateDataCallCount()).To(Equal(0), "Should not call GetPrivateData")
				key := stub.GetStateArgsForCall(0)
				Expect(key).To(Equal("007"), "Should call GetState with correct key")

				Expect(stub.PutStateCallCount()).To(Equal(0), "Should not call PutState")
				Expect(stub.PutPrivateDataCallCount()).To(Equal(0), "Should not call PutPrivateData")
			})
		})

		Context("With a named collection CreateState request", func() {
			var payload []byte

			BeforeEach(func() {
				context := &contract.TransactionContext{}
				context.ChannelId = "channel1"
				context.TransactionId = "txn1"
				collection := &contract.Collection{}
				collection.Name = "private"
				state := &contract.State{}
				state.Key = "007"
				state.Value = []byte("bond")
				request := &contract.CreateStateRequest{}
				request.Context = context
				request.Collection = collection
				request.State = state
				payload, _ = proto.Marshal(request)
			})

			It("should handle missing context error", func() {
				result, err := proxy.FabricCall(ctx, "wapc", "LedgerService", "CreateState", payload)
				Expect(result).To(BeNil())
				Expect(err).To(MatchError("CreateState failed: No stub found for transaction context channel1 txn1"))
			})

			It("should create a new state if it does not already exist in a named collection", func() {
				stub := &fakes.ChaincodeStubInterface{}
				contextStore.Put("channel1", "txn1", stub)

				Expect(proxy.FabricCall(ctx, "wapc", "LedgerService", "CreateState", payload)).To(BeNil())

				Expect(stub.GetStateCallCount()).To(Equal(0), "Should not call GetState")
				Expect(stub.GetPrivateDataCallCount()).To(Equal(1), "Should call GetPrivateData once")
				collection, key := stub.GetPrivateDataArgsForCall(0)
				Expect(collection).To(Equal("private"), "Should call GetPrivateData with correct collection name")
				Expect(key).To(Equal("007"), "Should call GetPrivateData with correct key")

				Expect(stub.PutStateCallCount()).To(Equal(0), "Should not call PutState")
				Expect(stub.PutPrivateDataCallCount()).To(Equal(1), "Should call PutPrivateData once")
				collection, key, value := stub.PutPrivateDataArgsForCall(0)
				Expect(collection).To(Equal("private"), "Should call PutPrivateData with correct collection name")
				Expect(key).To(Equal("007"), "Should call PutPrivateData with correct key")
				Expect(value).To(Equal([]byte("bond")), "Should call PutPrivateData with correct value")
			})

			It("should fail if the state already exists in a named collection", func() {
				stub := &fakes.ChaincodeStubInterface{}
				stub.GetPrivateDataReturns([]byte("dr evil"), nil)
				contextStore.Put("channel1", "txn1", stub)

				result, err := proxy.FabricCall(ctx, "wapc", "LedgerService", "CreateState", payload)
				Expect(result).To(BeNil())
				Expect(err).To(MatchError("CreateState failed for collection private: State already exists for key 007"))

				Expect(stub.GetStateCallCount()).To(Equal(0), "Should not call GetState")
				Expect(stub.GetPrivateDataCallCount()).To(Equal(1), "Should call GetPrivateData once")
				collection, key := stub.GetPrivateDataArgsForCall(0)
				Expect(collection).To(Equal("private"), "Should call GetPrivateData with correct collection name")
				Expect(key).To(Equal("007"), "Should call GetPrivateData with correct key")

				Expect(stub.PutStateCallCount()).To(Equal(0), "Should not call PutState")
				Expect(stub.PutPrivateDataCallCount()).To(Equal(0), "Should not call PutPrivateData")
			})
		})

		Context("With a world state ReadState request", func() {
			var payload []byte

			BeforeEach(func() {
				context := &contract.TransactionContext{}
				context.ChannelId = "channel1"
				context.TransactionId = "txn1"
				request := &contract.ReadStateRequest{}
				request.Context = context
				request.StateKey = "007"
				payload, _ = proto.Marshal(request)
			})

			It("should handle missing context error", func() {
				result, err := proxy.FabricCall(ctx, "wapc", "LedgerService", "ReadState", payload)
				Expect(result).To(BeNil())
				Expect(err).To(MatchError("ReadState failed: No stub found for transaction context channel1 txn1"))
			})

			It("should fail if the state key does not exist in the world state", func() {
				stub := &fakes.ChaincodeStubInterface{}
				contextStore.Put("channel1", "txn1", stub)
				stub.GetStateReturns(nil, nil)

				result, err := proxy.FabricCall(ctx, "wapc", "LedgerService", "ReadState", payload)
				Expect(result).To(BeNil())
				Expect(err).To(MatchError("ReadState failed: State 007 does not exist"))

				Expect(stub.GetStateCallCount()).To(Equal(1), "Should call GetState once")
				Expect(stub.GetPrivateDataCallCount()).To(Equal(0), "Should not call GetPrivateData")
				key := stub.GetStateArgsForCall(0)
				Expect(key).To(Equal("007"), "Should call GetState with correct key")
			})

			It("should return the correct value if the state key does exist in the world state", func() {
				stub := &fakes.ChaincodeStubInterface{}
				contextStore.Put("channel1", "txn1", stub)
				stub.GetStateReturns([]byte("bond"), nil)

				result, err := proxy.FabricCall(ctx, "wapc", "LedgerService", "ReadState", payload)
				Expect(err).To(BeNil())

				Expect(stub.GetStateCallCount()).To(Equal(1), "Should call GetState once")
				Expect(stub.GetPrivateDataCallCount()).To(Equal(0), "Should not call GetPrivateData")
				key := stub.GetStateArgsForCall(0)
				Expect(key).To(Equal("007"), "Should call GetState with correct key")

				response := &contract.ReadStateResponse{}
				_ = proto.Unmarshal(result, response)
				state := response.GetState()
				Expect(state.Key).To(Equal("007"))
				Expect(state.Value).To(Equal([]byte("bond")))
			})
		})

		Context("With a named collection ReadState request", func() {
			var payload []byte

			BeforeEach(func() {
				context := &contract.TransactionContext{}
				context.ChannelId = "channel1"
				context.TransactionId = "txn1"
				collection := &contract.Collection{}
				collection.Name = "private"
				request := &contract.ReadStateRequest{}
				request.Context = context
				request.Collection = collection
				request.StateKey = "007"
				payload, _ = proto.Marshal(request)
			})

			It("should handle missing context error", func() {
				result, err := proxy.FabricCall(ctx, "wapc", "LedgerService", "ReadState", payload)
				Expect(result).To(BeNil())
				Expect(err).To(MatchError("ReadState failed: No stub found for transaction context channel1 txn1"))
			})

			It("should fail if the state key does not exist in a named collection", func() {
				stub := &fakes.ChaincodeStubInterface{}
				contextStore.Put("channel1", "txn1", stub)
				stub.GetStateReturns(nil, nil)

				result, err := proxy.FabricCall(ctx, "wapc", "LedgerService", "ReadState", payload)
				Expect(result).To(BeNil())
				Expect(err).To(MatchError("ReadState failed for collection private: State 007 does not exist"))

				Expect(stub.GetStateCallCount()).To(Equal(0), "Should not call GetState")
				Expect(stub.GetPrivateDataCallCount()).To(Equal(1), "Should call GetPrivateData once")
				collection, key := stub.GetPrivateDataArgsForCall(0)
				Expect(collection).To(Equal("private"), "Should call GetPrivateData with correct collection name")
				Expect(key).To(Equal("007"), "Should call GetPrivateData with correct key")
			})

			It("should return the correct value if the state key does exist in a named collection", func() {
				stub := &fakes.ChaincodeStubInterface{}
				contextStore.Put("channel1", "txn1", stub)
				stub.GetPrivateDataReturns([]byte("bond"), nil)

				result, err := proxy.FabricCall(ctx, "wapc", "LedgerService", "ReadState", payload)
				Expect(err).To(BeNil())

				Expect(stub.GetStateCallCount()).To(Equal(0), "Should not call GetState")
				Expect(stub.GetPrivateDataCallCount()).To(Equal(1), "Should call GetPrivateData once")
				collection, key := stub.GetPrivateDataArgsForCall(0)
				Expect(collection).To(Equal("private"), "Should call GetPrivateData with correct collection name")
				Expect(key).To(Equal("007"), "Should call GetPrivateData with correct key")

				response := &contract.ReadStateResponse{}
				_ = proto.Unmarshal(result, response)
				state := response.GetState()
				Expect(state.Key).To(Equal("007"))
				Expect(state.Value).To(Equal([]byte("bond")))
			})
		})

		Context("With a world state ExistsState request", func() {
			var payload []byte

			BeforeEach(func() {
				context := &contract.TransactionContext{}
				context.ChannelId = "channel1"
				context.TransactionId = "txn1"
				request := &contract.ExistsStateRequest{}
				request.Context = context
				request.StateKey = "007"
				payload, _ = proto.Marshal(request)
			})

			It("should handle missing context error", func() {
				result, err := proxy.FabricCall(ctx, "wapc", "LedgerService", "ExistsState", payload)
				Expect(result).To(BeNil())
				Expect(err).To(MatchError("ExistsState failed: No stub found for transaction context channel1 txn1"))
			})

			It("should return false if the state key does not exist in the world state", func() {
				stub := &fakes.ChaincodeStubInterface{}
				contextStore.Put("channel1", "txn1", stub)
				stub.GetStateReturns(nil, nil)

				result, err := proxy.FabricCall(ctx, "wapc", "LedgerService", "ExistsState", payload)
				Expect(err).To(BeNil())

				Expect(stub.GetStateCallCount()).To(Equal(1), "Should call GetState once")
				Expect(stub.GetPrivateDataCallCount()).To(Equal(0), "Should not call GetPrivateData")
				key := stub.GetStateArgsForCall(0)
				Expect(key).To(Equal("007"), "Should call GetState with correct key")

				response := &contract.ExistsStateResponse{}
				_ = proto.Unmarshal(result, response)
				Expect(response.GetExists()).To(BeFalse())
			})

			It("should return true if the state key does exist in the world state", func() {
				stub := &fakes.ChaincodeStubInterface{}
				contextStore.Put("channel1", "txn1", stub)
				stub.GetStateReturns([]byte("bond"), nil)

				result, err := proxy.FabricCall(ctx, "wapc", "LedgerService", "ExistsState", payload)
				Expect(err).To(BeNil())

				Expect(stub.GetStateCallCount()).To(Equal(1), "Should call GetState once")
				Expect(stub.GetPrivateDataCallCount()).To(Equal(0), "Should not call GetPrivateData")
				key := stub.GetStateArgsForCall(0)
				Expect(key).To(Equal("007"), "Should call GetState with correct key")

				response := &contract.ExistsStateResponse{}
				_ = proto.Unmarshal(result, response)
				Expect(response.GetExists()).To(BeTrue())
			})
		})

		Context("With a named collection ExistsState request", func() {
			var payload []byte

			BeforeEach(func() {
				context := &contract.TransactionContext{}
				context.ChannelId = "channel1"
				context.TransactionId = "txn1"
				collection := &contract.Collection{}
				collection.Name = "private"
				request := &contract.ExistsStateRequest{}
				request.Context = context
				request.Collection = collection
				request.StateKey = "007"
				payload, _ = proto.Marshal(request)
			})

			It("should handle missing context error", func() {
				result, err := proxy.FabricCall(ctx, "wapc", "LedgerService", "ExistsState", payload)
				Expect(result).To(BeNil())
				Expect(err).To(MatchError("ExistsState failed: No stub found for transaction context channel1 txn1"))
			})

			It("should return false if the state key does not exist in the world state", func() {
				stub := &fakes.ChaincodeStubInterface{}
				contextStore.Put("channel1", "txn1", stub)
				stub.GetPrivateDataReturns(nil, nil)

				result, err := proxy.FabricCall(ctx, "wapc", "LedgerService", "ExistsState", payload)
				Expect(err).To(BeNil())

				Expect(stub.GetStateCallCount()).To(Equal(0), "Should not call GetState")
				Expect(stub.GetPrivateDataCallCount()).To(Equal(1), "Should call GetPrivateData once")
				collection, key := stub.GetPrivateDataArgsForCall(0)
				Expect(collection).To(Equal("private"), "Should call GetPrivateData with correct collection name")
				Expect(key).To(Equal("007"), "Should call GetPrivateData with correct key")

				response := &contract.ExistsStateResponse{}
				_ = proto.Unmarshal(result, response)
				Expect(response.GetExists()).To(BeFalse())
			})

			It("should return true if the state key does exist in the world state", func() {
				stub := &fakes.ChaincodeStubInterface{}
				contextStore.Put("channel1", "txn1", stub)
				stub.GetPrivateDataReturns([]byte("bond"), nil)

				result, err := proxy.FabricCall(ctx, "wapc", "LedgerService", "ExistsState", payload)
				Expect(err).To(BeNil())

				Expect(stub.GetStateCallCount()).To(Equal(0), "Should not call GetState")
				Expect(stub.GetPrivateDataCallCount()).To(Equal(1), "Should call GetPrivateData once")
				collection, key := stub.GetPrivateDataArgsForCall(0)
				Expect(collection).To(Equal("private"), "Should call GetPrivateData with correct collection name")
				Expect(key).To(Equal("007"), "Should call GetPrivateData with correct key")

				response := &contract.ExistsStateResponse{}
				_ = proto.Unmarshal(result, response)
				Expect(response.GetExists()).To(BeTrue())
			})
		})

		Context("With a world state UpdateState request", func() {
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

			It("should fail if the state does not already exist in the world state", func() {
				stub := &fakes.ChaincodeStubInterface{}
				contextStore.Put("channel1", "txn1", stub)

				result, err := proxy.FabricCall(ctx, "wapc", "LedgerService", "UpdateState", payload)
				Expect(result).To(BeNil())
				Expect(err).To(MatchError("UpdateState failed: No state exists for key 007"))

				Expect(stub.GetStateCallCount()).To(Equal(1), "Should call GetState once")
				Expect(stub.GetPrivateDataCallCount()).To(Equal(0), "Should not call GetPrivateData")
				key := stub.GetStateArgsForCall(0)
				Expect(key).To(Equal("007"), "Should call GetState with correct key")

				Expect(stub.PutStateCallCount()).To(Equal(0), "Should not call PutState")
				Expect(stub.PutPrivateDataCallCount()).To(Equal(0), "Should not call PutPrivateData")
			})

			It("should update a state which already exists in the world state", func() {
				stub := &fakes.ChaincodeStubInterface{}
				stub.GetStateReturns([]byte("dr evil"), nil)
				contextStore.Put("channel1", "txn1", stub)

				Expect(proxy.FabricCall(ctx, "wapc", "LedgerService", "UpdateState", payload)).To(BeNil())

				Expect(stub.GetStateCallCount()).To(Equal(1), "Should call GetState once")
				Expect(stub.GetPrivateDataCallCount()).To(Equal(0), "Should not call GetPrivateData")
				key := stub.GetStateArgsForCall(0)
				Expect(key).To(Equal("007"), "Should call GetState with correct key")

				Expect(stub.PutStateCallCount()).To(Equal(1), "Should call PutState once")
				Expect(stub.PutPrivateDataCallCount()).To(Equal(0), "Should not call PutPrivateData")
				key, value := stub.PutStateArgsForCall(0)
				Expect(key).To(Equal("007"), "Should call PutState with correct key")
				Expect(value).To(Equal([]byte("bond")), "Should call PutState with correct value")
			})
		})

		Context("With a named collection UpdateState request", func() {
			var payload []byte

			BeforeEach(func() {
				context := &contract.TransactionContext{}
				context.ChannelId = "channel1"
				context.TransactionId = "txn1"
				collection := &contract.Collection{}
				collection.Name = "private"
				state := &contract.State{}
				state.Key = "007"
				state.Value = []byte("bond")
				request := &contract.UpdateStateRequest{}
				request.Context = context
				request.Collection = collection
				request.State = state
				payload, _ = proto.Marshal(request)
			})

			It("should handle missing context error", func() {
				result, err := proxy.FabricCall(ctx, "wapc", "LedgerService", "UpdateState", payload)
				Expect(result).To(BeNil())
				Expect(err).To(MatchError("UpdateState failed: No stub found for transaction context channel1 txn1"))
			})

			It("should fail if the state does not already exist in a named collection", func() {
				stub := &fakes.ChaincodeStubInterface{}
				contextStore.Put("channel1", "txn1", stub)

				result, err := proxy.FabricCall(ctx, "wapc", "LedgerService", "UpdateState", payload)
				Expect(result).To(BeNil())
				Expect(err).To(MatchError("UpdateState failed for collection private: No state exists for key 007"))

				Expect(stub.GetStateCallCount()).To(Equal(0), "Should not call GetState")
				Expect(stub.GetPrivateDataCallCount()).To(Equal(1), "Should not call GetPrivateData once")
				collection, key := stub.GetPrivateDataArgsForCall(0)
				Expect(collection).To(Equal("private"), "Should call GetPrivateData with correct collection name")
				Expect(key).To(Equal("007"), "Should call GetState with correct key")

				Expect(stub.PutStateCallCount()).To(Equal(0), "Should not call PutState")
				Expect(stub.PutPrivateDataCallCount()).To(Equal(0), "Should not call PutPrivateData")
			})

			It("should update a state which already exists in a named collection", func() {
				stub := &fakes.ChaincodeStubInterface{}
				stub.GetPrivateDataReturns([]byte("dr evil"), nil)
				contextStore.Put("channel1", "txn1", stub)

				Expect(proxy.FabricCall(ctx, "wapc", "LedgerService", "UpdateState", payload)).To(BeNil())

				Expect(stub.GetStateCallCount()).To(Equal(0), "Should not call GetState")
				Expect(stub.GetPrivateDataCallCount()).To(Equal(1), "Should not call GetPrivateData once")
				collection, key := stub.GetPrivateDataArgsForCall(0)
				Expect(collection).To(Equal("private"), "Should call GetPrivateData with correct collection name")
				Expect(key).To(Equal("007"), "Should call GetState with correct key")

				Expect(stub.PutStateCallCount()).To(Equal(0), "Should not call PutState")
				Expect(stub.PutPrivateDataCallCount()).To(Equal(1), "Should call PutPrivateData once")
				collection, key, value := stub.PutPrivateDataArgsForCall(0)
				Expect(collection).To(Equal("private"), "Should call PutPrivateData with correct collection name")
				Expect(key).To(Equal("007"), "Should call PutPrivateData with correct key")
				Expect(value).To(Equal([]byte("bond")), "Should call PutPrivateData with correct value")
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

				result, err := proxy.FabricCall(ctx, "wapc", "LedgerService", "GetStates", payload)
				Expect(err).To(BeNil())

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
