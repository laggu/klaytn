package grpc

import "ground-x/go-gxplatform/accounts/abi"

// Parse gRPC methods and required message types from methods in an Ethereum contract ABI.
func ParseMethods(abiMethods map[string]abi.Method) (methods Methods, msgs []Message) {
	for _, f := range abiMethods {
		method, msg := ParseMethod(f)
		methods = append(methods, method)
		msgs = append(msgs, msg...)
	}

	return
}

// Parse gRPC method and required message types from an Ethereum contract method.
func ParseMethod(m abi.Method) (Method, []Message) {
	method := Method{
		Const: m.Const,
		Name:  m.Name,
	}

	method.Inputs = append(method.Inputs, parseArgs(m.Inputs)...)
	method.Outputs = append(method.Outputs, parseArgs(m.Outputs)...)

	// If it is not a const method, we need to provide
	// more transaction options to send transactions.
	if !m.Const {
		method.Inputs = append(method.Inputs, Argument{
			Name:    "opts",
			Type:    TransactOptsReq.Name,
			IsSlice: false,
		})
	}

	var requiredMessages []Message
	if len(method.Inputs) > 0 {
		requiredMessages = append(requiredMessages, ToMessage(method.RequestName(), method.Inputs))
	}
	if len(method.Outputs) > 0 {
		requiredMessages = append(requiredMessages, ToMessage(method.ResponseName(), method.Outputs))
	}

	return method, requiredMessages
}

func parseArgs(args []abi.Argument) (results []Argument) {
	for _, arg := range args {
		results = append(results, ToGrpcArgument(arg))
	}

	return
}
