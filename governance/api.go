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

package governance

type PublicGovernanceAPI struct {
	governance *Governance // Node interfaced by this API
}

func NewGovernanceAPI(gov *Governance) *PublicGovernanceAPI {
	logger.Warn("In NewGovernanceAPI", "governance", gov)
	return &PublicGovernanceAPI{governance: gov}
}

// Vote injects a new vote for governance targets such as unitprice and governingnode.
func (api *PublicGovernanceAPI) Vote(key string, val interface{}) bool {
	logger.Warn("Address", "governance", api.governance)
	return api.governance.AddVote(key, val)
}
