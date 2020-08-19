// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package internal_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/hyperledgendary/fabric-chaincode-wasm/internal"
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
			It("should error if the binding is not wasm", func() {
				result, err := proxy.FabricCall(ctx, "notWasm", "LedgerService", "CreateState", []byte(""))
				Expect(result).To(BeNil())
				Expect(err).To(MatchError("Operation not supported: notWasm LedgerService CreateState"))
			})

			It("should error if the namespace is not LedgerService", func() {
				result, err := proxy.FabricCall(ctx, "wasm", "TeaService", "CreateState", []byte(""))
				Expect(result).To(BeNil())
				Expect(err).To(MatchError("Operation not supported: wasm TeaService CreateState"))
			})

			It("should error with an unsupported operation", func() {
				result, err := proxy.FabricCall(ctx, "wasm", "LedgerService", "MutateState", []byte(""))
				Expect(result).To(BeNil())
				Expect(err).To(MatchError("Operation not supported: wasm LedgerService MutateState"))
			})
		})
	})

})
