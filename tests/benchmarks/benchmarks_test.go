// Copyright 2018 The go-klaytn Authors
//
// This file is part of the go-klaytn library.
//
// The go-klaytn library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-klaytn library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-klaytn library. If not, see <http://www.gnu.org/licenses/>.

package benchmarks

import (
	"testing"

	"github.com/ground-x/go-gxplatform/common"
)


func BenchmarkInterpreterMload100000(bench *testing.B) {
	//
	// Test code
	//       Initialize memory with memory write (PUSH PUSH MSTORE)
	//       Loop 10000 times for below code
	//              memory read 10 times //  (PUSH MLOAD POP) x 10
	//
	code := common.Hex2Bytes("60ca60205260005b612710811015630000004557602051506020515060205150602051506020515060205150602051506020515060205150602051506001016300000007565b00")
	intrp, contract := prepareInterpreterAndContract(code)

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		intrp.Run(contract, nil)
	}
}

func BenchmarkInterpreterMstore100000(bench *testing.B) {
	//
	// Test code
	//       Initialize memory with memory write (PUSH PUSH MSTORE)
	//       Loop 10000 times for below code
	//              memory write 10 times //  (PUSH PUSH MSTORE) x 10
	//
	code := common.Hex2Bytes("60ca60205260005b612710811015630000004f5760fe60205260fe60205260fe60205260fe60205260fe60205260fe60205260fe60205260fe60205260fe60205260fe6020526001016300000007565b00")
	intrp, contract := prepareInterpreterAndContract(code)

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		intrp.Run(contract, nil)
	}
}

func BenchmarkInterpreterSload100000(bench *testing.B) {
	//
	// Test code
	//       Initialize (PUSH SSTORE)
	//       Loop 10000 times for below code
	//              Read from storage 10 times //  (PUSH SLOAD POP) x 10
	//
	code := common.Hex2Bytes("60ca60205560005b612710811015630000004557602054506020545060205450602054506020545060205450602054506020545060205450602054506001016300000007565b00")
	intrp, contract := prepareInterpreterAndContract(code)

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		intrp.Run(contract, nil)
	}
}

func BenchmarkInterpreterSstore100000(bench *testing.B) {
	//
	// Test code
	//       Initialize (PUSH)
	//       Loop 10000 times for below code
	//              Write to storage 10 times //  (PUSH PUSH SSTORE) x 10
	//
	code := common.Hex2Bytes("60005b612710811015630000004a5760fe60205560fe60205560fe60205560fe60205560fe60205560fe60205560fe60205560fe60205560fe60205560fe6020556001016300000002565b00")
	intrp, contract := prepareInterpreterAndContract(code)

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		intrp.Run(contract, nil)
	}
}

func BenchmarkInterpreterAdd100000(bench *testing.B) {
	//
	// Test code
	//       Initialize value (PUSH)
	//       Loop 10000 times for below code
	//              initialize value and increment 10 times //  (PUSH x 1) + ((PUSH ADD) x 10)
	//
	code := common.Hex2Bytes("60ca60205260005b612710811015630000003e576000600101600101600101600101600101600101600101600101600101600101506001016300000007565b00")
	intrp, contract := prepareInterpreterAndContract(code)

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		intrp.Run(contract, nil)
	}
}

func BenchmarkInterpreterPush1Mul1byte100000(bench *testing.B) {
	//
	// Test code
	//       Initialize value (PUSH)
	//       Loop 10000 times for below code
	//              initialize value and increment 10 times //  (PUSH1 PUSH1 MUL POP) x 10
	//
	code := common.Hex2Bytes("60005b61271081101563000000545760ca60fe025060ca60fe025060ca60fe025060ca60fe025060ca60fe025060ca60fe025060ca60fe025060ca60fe025060ca60fe025060ca60fe02506001016300000002565b00")
	intrp, contract := prepareInterpreterAndContract(code)

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		intrp.Run(contract, nil)
	}
}

func BenchmarkInterpreterPush5Mul1byte100000(bench *testing.B) {
	//
	// Test code
	//       Initialize value (PUSH)
	//       Loop 10000 times for below code
	//              initialize value and increment 10 times //  (PUSH5 PUSH5 MUL POP) x 10
	//
	code := common.Hex2Bytes("60005b61271081101563000000a4576400000000ca6400000000fe02506400000000ca6400000000fe02506400000000ca6400000000fe02506400000000ca6400000000fe02506400000000ca6400000000fe02506400000000ca6400000000fe02506400000000ca6400000000fe02506400000000ca6400000000fe02506400000000ca6400000000fe02506400000000ca6400000000fe02506001016300000002565b00")
	intrp, contract := prepareInterpreterAndContract(code)

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		intrp.Run(contract, nil)
	}
}

func BenchmarkInterpreterPush5Mul5bytes100000(bench *testing.B) {
	//
	// Test code
	//       Initialize value (PUSH)
	//       Loop 10000 times for below code
	//              initialize value and increment 10 times //  (PUSH5 PUSH5 MUL POP) x 10
	//
	code := common.Hex2Bytes("60005b61271081101563000000a45764cafebabe0064caffe00000025064cafebabe0064caffe00000025064cafebabe0064caffe00000025064cafebabe0064caffe00000025064cafebabe0064caffe00000025064cafebabe0064caffe00000025064cafebabe0064caffe00000025064cafebabe0064caffe00000025064cafebabe0064caffe00000025064cafebabe0064caffe0000002506001016300000002565b00")
	intrp, contract := prepareInterpreterAndContract(code)

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		intrp.Run(contract, nil)
	}
}

func BenchmarkInterpreterPush1Div1byte100000(bench *testing.B) {
	//
	// Test code
	//       Initialize value (PUSH)
	//       Loop 10000 times for below code
	//              initialize value and increment 10 times //  (PUSH1 PUSH1 DIV POP) x 10
	//
	code := common.Hex2Bytes("60005b61271081101563000000545760ca60fe045060ca60fe045060ca60fe045060ca60fe045060ca60fe045060ca60fe045060ca60fe045060ca60fe045060ca60fe045060ca60fe04506001016300000002565b00")
	intrp, contract := prepareInterpreterAndContract(code)

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		intrp.Run(contract, nil)
	}
}

func BenchmarkInterpreterPush5Div1byte100000(bench *testing.B) {
	//
	// Test code
	//       Initialize value (PUSH)
	//       Loop 10000 times for below code
	//              initialize value and increment 10 times //  (PUSH5 PUSH5 DIV POP) x 10
	//
	code := common.Hex2Bytes("60005b61271081101563000000a4576400000000ca6400000000fe04506400000000ca6400000000fe04506400000000ca6400000000fe04506400000000ca6400000000fe04506400000000ca6400000000fe04506400000000ca6400000000fe04506400000000ca6400000000fe04506400000000ca6400000000fe04506400000000ca6400000000fe04506400000000ca6400000000fe04506001016300000002565b00")
	intrp, contract := prepareInterpreterAndContract(code)

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		intrp.Run(contract, nil)
	}
}

func BenchmarkInterpreterPush5Div5bytes100000(bench *testing.B) {
	//
	// Test code
	//       Initialize value (PUSH)
	//       Loop 10000 times for below code
	//              initialize value and increment 10 times //  (PUSH5 PUSH5 DIV POP) x 10
	//
	code := common.Hex2Bytes("60005b61271081101563000000a45764cafebabe0064caffe00000045064cafebabe0064caffe00000045064cafebabe0064caffe00000045064cafebabe0064caffe00000045064cafebabe0064caffe00000045064cafebabe0064caffe00000045064cafebabe0064caffe00000045064cafebabe0064caffe00000045064cafebabe0064caffe00000045064cafebabe0064caffe0000004506001016300000002565b00")
	intrp, contract := prepareInterpreterAndContract(code)

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		intrp.Run(contract, nil)
	}
}
