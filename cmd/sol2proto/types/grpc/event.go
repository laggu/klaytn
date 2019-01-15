package grpc

import (
	"github.com/ground-x/klaytn/accounts/abi"
	"github.com/ground-x/klaytn/cmd/utils"
)

// Parse gRPC methods and required message types from events in an Ethereum contract ABI.
func ParseEvents(abiEvents map[string]abi.Event) (methods Methods, msgs []Message) {
	for _, ev := range abiEvents {
		method, msg := ParseEvent(ev)
		methods = append(methods, method)
		msgs = append(msgs, msg...)
	}

	return
}

// Parse gRPC method and required message types from an Ethereum event.
func ParseEvent(ev abi.Event) (Method, []Message) {
	method := Method{}

	if ev.Anonymous {
		method.Name = "onEvent" + utils.ToCamelCase(ev.Id().Hex())
	} else {
		method.Name = "on" + ev.Name
	}

	method.Inputs = append(method.Inputs, parseArgs(ev.Inputs)...)

	var requiredMessages []Message
	if len(method.Inputs) > 0 {
		requiredMessages = append(requiredMessages, ToMessage(method.RequestName(), method.Inputs))
	}

	return method, requiredMessages
}
