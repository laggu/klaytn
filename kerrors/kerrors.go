// Copyright 2018 The klaytn Authors
// This file is part of the klaytn library.
//
// The klaytn library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The klaytn library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the klaytn library. If not, see <http://www.gnu.org/licenses/>.

package kerrors

import "errors"

// TODO-Klaytn: Use integer for error codes.
// TODO-Klaytn: Integrate all universally accessible errors into kerrors package.
var (
	ErrOutOfGas                  = errors.New("out of gas")
	ErrMaxKeysExceed             = errors.New("the number of keys exceeds the limit")
	ErrMaxKeysExceedInValidation = errors.New("the number of keys exceeds the limit in the validation check")
	ErrMaxFeeRatioExceeded       = errors.New("fee ratio exceeded the maximum")
	ErrEmptySlice                = errors.New("slice is empty")
	ErrNotProgramAccount         = errors.New("not a program account (e.g., an account having code and storage)")
)
