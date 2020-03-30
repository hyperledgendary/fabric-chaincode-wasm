package wasmruntime

import (
	"fmt"
	"log"

	datatypes "wcr/datatypes"

	"github.com/golang/protobuf/proto"
	"github.com/perlin-network/life/exec"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
)

// WasmPcRuntime is an abstraction of the instance of the Wasm engine
type WasmPcRuntime struct {
	guestCall int
	callctx   Callcontext
	vm        *exec.VirtualMachine
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Callcontext Represents an individual call
//
type Callcontext struct {
	id     string
	fnName string
	buffer []byte

	hostCallResult []byte

	finalResult *datatypes.Return
}

// Resolver defines the imports to the Wasm code
type Resolver struct {
	runtime *WasmPcRuntime
}

// ResolveFunc Sorts out the imports of the runtime
func (r *Resolver) ResolveFunc(module, field string) exec.FunctionImport {
	// log.Printf("Resolve func: %s %s\n", module, field)
	wr := *(r.runtime)
	switch module {
	case "wapc":
		switch field {
		case "__guest_request":
			log.Printf("__guest_request txid=%s", (wr.callctx.id))
			return func(vm *exec.VirtualMachine) int64 {
				//Pointer and length for key
				opptr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				paramptr := int(uint32(vm.GetCurrentFrame().Locals[1]))

				fnBuffer := []byte("contract") // []byte(wr.callctx.fnName)
				m := vm.Memory[opptr : opptr+len(fnBuffer)]
				copy(m, fnBuffer)

				m = vm.Memory[paramptr : paramptr+len(wr.callctx.buffer)]
				copy(m, wr.callctx.buffer)

				return 1
			}
		case "__guest_error":
			log.Printf("__guest_error txid=%s", (wr.callctx.id))
			return func(vm *exec.VirtualMachine) int64 {
				ptr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				len := int(uint32(vm.GetCurrentFrame().Locals[1]))
				m := vm.Memory[ptr : ptr+len]
				log.Printf("%x", m)
				// do something with m, result (but in error) buffer
				return 1
			}
		case "__guest_response":
			log.Printf("__guest_response txid=%s", (wr.callctx.id))
			return func(vm *exec.VirtualMachine) int64 {
				ptr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				len := int(uint32(vm.GetCurrentFrame().Locals[1]))
				m := vm.Memory[ptr : ptr+len]
				log.Printf("memory at %x", m)

				// take the buffer and inflate from protobuf

				err := proto.Unmarshal(m, wr.callctx.finalResult)
				if err != nil {
					log.Fatal("unmarshalling error ", err)
					return 0
				}
				log.Printf("protobuf is %s", wr.callctx.finalResult)

				return 1
			}
		case "__host_call":
			// ns_ptr, ns_len, op_ptr, op_len, prt, len
			log.Printf("__host_call txid=%s", (wr.callctx.id))
			return func(vm *exec.VirtualMachine) int64 {
				nsptr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				nslen := int(uint32(vm.GetCurrentFrame().Locals[1]))
				namespace := string(vm.Memory[nsptr : nsptr+nslen])

				opptr := int(uint32(vm.GetCurrentFrame().Locals[2]))
				oplen := int(uint32(vm.GetCurrentFrame().Locals[3]))
				op := string(vm.Memory[opptr : opptr+oplen])

				ptr := int(uint32(vm.GetCurrentFrame().Locals[4]))
				len := int(uint32(vm.GetCurrentFrame().Locals[5]))
				m := vm.Memory[ptr : ptr+len]
				// do something with m, result buffer
				log.Printf("ns %x %x %s", nsptr, nslen, namespace)
				log.Printf("op %x %x %s", opptr, oplen, op)
				log.Printf("mm %x %x %s", ptr, len, m)
				// make the actual call to the function
				// will pretend by setting the result with a string
				wr.callctx.hostCallResult = []byte("Hello from Go")
				return 1 // SUCCESS
			}
		case "__host_response_len":
			log.Printf("__host_response_len txid=%s", (wr.callctx.id))
			return func(vm *exec.VirtualMachine) int64 {
				return int64(len(wr.callctx.hostCallResult))
			}
		case "__host_response":
			// ptr to the location to copy the data
			log.Printf("__host_response txid=%s", (wr.callctx.id))
			return func(vm *exec.VirtualMachine) int64 {
				ptr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				m := vm.Memory[ptr : ptr+len(wr.callctx.hostCallResult)]
				copy(m, wr.callctx.hostCallResult)

				return 1
			}
		case "__host_error_len":
			// treat the error response as pure response
			log.Printf("__host_error_len txid=%s", (wr.callctx.id))
			return func(vm *exec.VirtualMachine) int64 {
				return int64(len(wr.callctx.hostCallResult))
			}
		case "__host_error":
			log.Printf("__host_error txid=%s", (wr.callctx.id))
			return func(vm *exec.VirtualMachine) int64 {
				ptr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				m := vm.Memory[ptr : ptr+len(wr.callctx.hostCallResult)]
				copy(m, wr.callctx.hostCallResult)

				return 1
			}
		case "__console_log":
			return func(vm *exec.VirtualMachine) int64 {
				ptr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				msgLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
				msg := vm.Memory[ptr : ptr+msgLen]

				log.Printf("[guest:console_log] %s\n", string(msg))

				return 0
			}
		case "__log":
			return func(vm *exec.VirtualMachine) int64 {
				ptr := int(uint32(vm.GetCurrentFrame().Locals[0]))
				msgLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
				msg := vm.Memory[ptr : ptr+msgLen]

				log.Printf("[guest:log] %s\n", string(msg))

				return 0
			}
		default:
			panic(fmt.Errorf("Unkown field: %s", module))
		}

	default:
		panic(fmt.Errorf("Unkown module: %s", module))
	}

}

// ResolveGlobal defines a set of global variables for use within a WebAssembly module.
func (r *Resolver) ResolveGlobal(module, field string) int64 {

	switch module {
	case "env":
		switch field {
		case "__constant_variable":
			return 424
		default:
			panic(fmt.Errorf("unknown field: %s", field))
		}
	default:
		panic(fmt.Errorf("unknown module: %s", module))
	}
}

// Call makes the requried tx call on the wasm contract
func (wr *WasmPcRuntime) Call(fnname string, args []string, txid string, channelid string) {

	fnBuffer := []byte("contract") //[]byte(fnname)
	wr.callctx = Callcontext{
		fnName: fnname,
	}
	wr.callctx.id = txid + ":" + channelid

	msg := &datatypes.Arguments{
		Fnname:    fnname,
		Txid:      txid,
		Channelid: channelid,
		Args:      args}
	wr.callctx.finalResult = &datatypes.Return{}

	argsBuffer, err := proto.Marshal(msg)
	check(err)
	wr.callctx.buffer = argsBuffer

	entryID, ok := wr.vm.GetFunctionExport("__guest_call")
	if !ok {
		panic("Entry function not found")
	}

	log.Printf("[host] calling %s txid=%s", fnname, txid)
	result, err := wr.vm.Run(entryID, int64(len(fnBuffer)), int64(len(argsBuffer)))
	log.Printf("%d %s", result, wr.callctx.finalResult.Data)
}

// NewRuntime Get the runtime
func NewRuntime(wasmBytes []byte) *WasmPcRuntime {

	wr := WasmPcRuntime{}

	resolver := Resolver{runtime: &wr}
	log.Printf("[host] New Wasm VM")
	// Instantiate a new WebAssembly VM with a few resolved imports.
	vm, err := exec.NewVirtualMachine(wasmBytes, exec.VMConfig{
		DefaultMemoryPages:   128,
		DefaultTableSize:     65536,
		DisableFloatingPoint: false,
	}, &resolver, nil)
	if err != nil {
		panic("can't create wasm engine")
	}
	wr.vm = vm

	log.Printf("[host] Main entry point  __guest_call")
	// Get the function ID of the entry function to be executed.
	entryID, ok := wr.vm.GetFunctionExport("__guest_call")
	if !ok {
		panic("Entry function not found")
	}

	wr.guestCall = entryID

	return &wr
}

func (s *WasmPcRuntime) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (s *WasmPcRuntime) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	function, args := APIstub.GetFunctionAndParameters()
	txid := APIstub.GetTxID()
	channelid := APIstub.GetChannelID()
	s.Call(function, args, txid, channelid)
	return shim.Success(nil)
}
