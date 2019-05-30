// Copyright 2019 The klaytn Authors
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

import "testing"

var (
	testStorage = &simpleStorage{}
	testData    = []*Node{
		MustParseNode("kni://2d2d43be39c40e1b104952cc351127fb3783b66ca065ba4b8c46f6e73e603e511203e399154c5b96c8ca13f8dd9086f6d29c74867a3b7ea6bd4f0205b25522b6@0.0.0.0:32323?discport=0"),
		MustParseNode("kni://e2fc0988b6286fd15a9c208bdf283fb456575357911d618e2f47326ce534db1f94c3a1edc6cb7a399797b35ad7dd82a2647ba9ec43b33948302cefb9edd2c9b2@0.0.0.0:32323?discport=0"),
		MustParseNode("kni://e898d53588ed46888ace5a6a61c2ca71034ae23aa004b8525a5045a5a51d43dd72b9ca49b346ed0155cc6e2cd143486109a6fe21845a59778012994ea7d9e128@0.0.0.0:32323?discport=0"),
		MustParseNode("kni://79ce147a6e955cbf43004e19480c4d8139eeb71ec99ee872499e4cb05e37ec711049781702200f958698246475e7bcf898acad261d1950c280abf5090ea00ac0@0.0.0.0:32323?discport=0"),
		MustParseNode("kni://693d678f8497a0a6019acdcc6388c489a0078387bb0dafa27e36b856765b20abe6a7c6e69b6bb4c0028babf490f7aea5481c2c39f69153c4f670af720ec59f67@0.0.0.0:32323?discport=0"),
		MustParseNode("kni://850555871dd0eeea187d7ad0f219b4def2a377cee8a6b40dff985b907687c4a36afa337a8499aff157ef41dadba8dcfb202786bb86d0a13371c8a517db3bfade@0.0.0.0:32323?discport=0"),
		MustParseNode("kni://dda3ce1adf19a1d88fba5fe91d856aa8c5a9fb71892b4cf75ffc19f28a6a38cc8eb3d4cdd83c2013ba8e8112fe44a43de0198e639171c4d5887452f0f7b21712@0.0.0.0:32323?discport=0"),
		MustParseNode("kni://473e2e359e2de9922eff61269a719004768c96733da823aaabbe6c926540f40156107b7fac931703d9f01a9cfc33652889c98fb62dbffe6a04a9ce2c93bc7512@0.0.0.0:32323?discport=0"),
		MustParseNode("kni://26a8bd88faadfe401787e99ca7909793cbe388562a0e43286488c2306770a73173ed0a50c5c894ae9e390da21c9091f4edf924d3e4299b303434628198b8a38a@0.0.0.0:32323?discport=0"),
		MustParseNode("kni://dbc67e54732ca531c8c3b2175858d712d1a1cac4a59ad6b639c6fe72bc261916c2236a38227502975f10e5353e4554dc50daaede6f8c9203beb36438af356d99@0.0.0.0:32323?discport=0"),
	}
)

func TestShuffle(t *testing.T) {
	// shuffle empty list
	testStorage.init()
	empty := testStorage.shuffle([]*Node{})
	if len(empty) != 0 {
		t.Errorf("the length of the shuffled empty list. expected: %v, actual: %v\n", 0, len(empty))
	}

	// shuffle an element list
	testStorage.init()
	oneElement := testStorage.shuffle([]*Node{testData[0]})
	if len(oneElement) != 1 {
		t.Errorf("the length of shuffled an element list. expected: %v, actual: %v\n", 1, len(oneElement))
	}
	if !oneElement[0].CompareNode(testData[0]) {
		t.Errorf("the shuffled result is wrong. expected: %v, acutal: %v\n", testData[0].String(), oneElement[0].String())
	}

	// shuffle the predefined list
	testStorage.init()
	list := testStorage.shuffle(testData)
	if len(list) != len(testData) {
		t.Errorf("the length of shuffled list is wrong. expected: %v, actual: %v\n", len(testData), len(list))
	}
	isOrderChanged := false
	for idx, shuffled := range list {
		if !testData[idx].CompareNode(shuffled) {
			isOrderChanged = true
		}
	}
	if !isOrderChanged {
		t.Error("the order of the list is not changed.")
	}
	for _, shuffled := range list {
		existed := false
		for _, original := range testData {
			if shuffled.CompareNode(original) {
				existed = true
			}
		}
		if !existed {
			t.Errorf("one of the elements does not exist after shuffling. missing: %v\n", shuffled.String())
		}
	}
}
