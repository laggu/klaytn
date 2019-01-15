package grpc

import "github.com/ground-x/klaytn/accounts/abi"

func ToGrpcArgument(in abi.Argument) Argument {
	arg := Argument{
		Name:    in.Name,
		IsSlice: in.Type.T == abi.SliceTy,
	}

	arg.Type = toGrpcType(in.Type)
	return arg
}

func toGrpcType(t abi.Type) string {
	switch t.T {
	case abi.IntTy:
		if t.Size == 8 {
			return "byte"
		} else if t.Size == 32 {
			return "int32"
		} else if t.Size == 64 {
			return "int64"
		}
		return "bytes"
	case abi.UintTy:
		if t.Size == 8 {
			return "byte"
		} else if t.Size == 32 {
			return "uint32"
		} else if t.Size == 64 {
			return "uint64"
		}
		return "bytes"
	case abi.BoolTy:
		return "bool"
	case abi.StringTy:
		return "string"
	case abi.AddressTy:
		return "string"
	case abi.FixedBytesTy:
		return "bytes"
	case abi.BytesTy:
		return "bytes"
	case abi.HashTy:
		return "string"
	case abi.FixedPointTy:
	case abi.FunctionTy:
		fallthrough
	default:
	}

	return "bytes"
}

func ToMessage(name string, args []Argument) Message {
	return Message{
		Name: name,
		Args: args,
	}
}
