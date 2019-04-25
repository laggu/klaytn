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

package discover

func NewDiscovery(cfg *Config) (Discovery, error) {
	switch cfg.DiscoveryPolicyPreset {
	case DiscoveryPolicyPresetPBN:
	case DiscoveryPolicyPresetPN:
	case DiscoveryPolicyPresetEN:
	case DiscoveryPolicyPresetCN:
	case DiscoveryPolicyPresetCBN:
		// TODO-Klaytn-Discovery: create simple discovery interface in here
	case DiscoveryPolicyPresetEBN:
	default:
		return newTable(cfg)
	}
	return newTable(cfg) // this is default interface creation for test codes
}
