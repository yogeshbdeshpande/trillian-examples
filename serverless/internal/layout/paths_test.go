// Copyright 2021 Google LLC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package layout

import (
	"fmt"
	"testing"
)

func TestSeqPath(t *testing.T) {
	for _, test := range []struct {
		root     string
		seq      uint64
		wantDir  string
		wantFile string
	}{
		{
			root:     "/root/path",
			seq:      0,
			wantDir:  "/root/path/seq/00/00/00/00",
			wantFile: "00",
		}, {
			root:     "/root/path",
			seq:      0x85,
			wantDir:  "/root/path/seq/00/00/00/00",
			wantFile: "85",
		}, {
			root:     "/a/different/root/path",
			seq:      0x86,
			wantDir:  "/a/different/root/path/seq/00/00/00/00",
			wantFile: "86",
		}, {
			root:     "/a/different/root/path",
			seq:      0xffeeddccbb,
			wantDir:  "/a/different/root/path/seq/ff/ee/dd/cc",
			wantFile: "bb",
		},
	} {
		desc := fmt.Sprintf("root %q seq %d", test.root, test.seq)
		t.Run(desc, func(t *testing.T) {
			gotDir, gotFile := SeqPath(test.root, test.seq)
			if gotDir != test.wantDir {
				t.Errorf("Got dir %q want %q", gotDir, test.wantDir)
			}
			if gotFile != test.wantFile {
				t.Errorf("got file %q want %q", gotFile, test.wantFile)
			}
		})
	}
}

func TestSeqFromPath(t *testing.T) {
	for _, test := range []struct {
		desc    string
		seqPath string
		root    string
		wantSeq uint64
		wantErr bool
	}{
		{
			desc:    "valid 0",
			root:    "/bananas",
			seqPath: "/bananas/seq/00/00/00/00/00",
			wantSeq: 0,
		}, {
			desc:    "valid x10",
			root:    "/bananas",
			seqPath: "/bananas/seq/00/00/00/00/10",
			wantSeq: 0x10,
		}, {
			desc:    "valid large",
			root:    "/lemons",
			seqPath: "/lemons/seq/12/34/56/78/90",
			wantSeq: 0x1234567890,
		}, {
			desc:    "wrong root",
			root:    "/lemons",
			seqPath: "/apples/seq/12/34/56/78/90",
			wantErr: true,
		}, {
			desc:    "missing seq path",
			root:    "/lemons",
			seqPath: "/lemons/12/34/56/78/90",
			wantErr: true,
		}, {
			desc:    "invalid format",
			root:    "/lemons",
			seqPath: "/lemons/12/4/56/78/90",
			wantErr: true,
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			got, gotErr := SeqFromPath(test.root, test.seqPath)
			if test.wantErr != (gotErr != nil) {
				t.Fatalf("Want err %t, got err %v", test.wantErr, gotErr)
			}
			if got != test.wantSeq {
				t.Errorf("got seq %d, want %d", got, test.wantSeq)
			}
		})
	}
}

func TestLeafPath(t *testing.T) {
	for _, test := range []struct {
		root     string
		hash     []byte
		wantDir  string
		wantFile string
	}{
		{
			root:     "/root/path",
			hash:     []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77},
			wantDir:  "/root/path/leaves/11/22/33",
			wantFile: "44556677",
		}, {
			root:     "/root/path",
			hash:     []byte{0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd},
			wantDir:  "/root/path/leaves/88/99/aa",
			wantFile: "bbccdd",
		}, {
			root:     "/a/different/root/path",
			hash:     []byte{0x12, 0x34, 0x56, 0x78, 0x9a},
			wantDir:  "/a/different/root/path/leaves/12/34/56",
			wantFile: "789a",
		},
	} {
		desc := fmt.Sprintf("root %q hash %x", test.root, test.hash)
		t.Run(desc, func(t *testing.T) {
			gotDir, gotFile := LeafPath(test.root, test.hash)
			if gotDir != test.wantDir {
				t.Errorf("Got dir %q want %q", gotDir, test.wantDir)
			}
			if gotFile != test.wantFile {
				t.Errorf("got file %q want %q", gotFile, test.wantFile)
			}
		})
	}
}

func TestTilePath(t *testing.T) {
	for _, test := range []struct {
		root     string
		level    uint64
		index    uint64
		tileSize uint64
		wantDir  string
		wantFile string
	}{
		{
			root:     "/root/path",
			level:    0,
			index:    0,
			tileSize: 0,
			wantDir:  "/root/path/tile/00/0000/00/00",
			wantFile: "00",
		}, {
			root:     "/root/path",
			level:    0,
			index:    0,
			tileSize: 1,
			wantDir:  "/root/path/tile/00/0000/00/00",
			wantFile: "00.01",
		}, {
			root:     "/root/path",
			level:    0x10,
			index:    0,
			wantDir:  "/root/path/tile/10/0000/00/00",
			wantFile: "00",
		}, {
			root:     "/root/path",
			level:    0x10,
			index:    0x455667,
			tileSize: 0x78,
			wantDir:  "/root/path/tile/10/0000/45/56",
			wantFile: "67.78",
		}, {
			root:     "/root/path",
			level:    0x10,
			index:    0x123456789a,
			tileSize: 0x7,
			wantDir:  "/root/path/tile/10/1234/56/78",
			wantFile: "9a.07",
		}, {
			root:     "/a/different/root/path",
			level:    0x15,
			index:    0x455667,
			wantDir:  "/a/different/root/path/tile/15/0000/45/56",
			wantFile: "67",
		},
	} {
		desc := fmt.Sprintf("root %q level %x index %x", test.root, test.level, test.index)
		t.Run(desc, func(t *testing.T) {
			gotDir, gotFile := TilePath(test.root, test.level, test.index, test.tileSize)
			if gotDir != test.wantDir {
				t.Errorf("Got dir %q want %q", gotDir, test.wantDir)
			}
			if gotFile != test.wantFile {
				t.Errorf("got file %q want %q", gotFile, test.wantFile)
			}
		})
	}
}
