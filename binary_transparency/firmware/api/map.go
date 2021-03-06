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

package api

import "fmt"

const (
	// MapHTTPGetCheckpoint is the path of the URL to get a recent map checkpoint.
	MapHTTPGetCheckpoint = "ftmap/v0/get-checkpoint"
	// MapHTTPGetTile is the path of the URL to get a map tile at a revision.
	MapHTTPGetTile = "ftmap/v0/tile"
	// MapHTTPGetAggregation is the path of the URL to get aggregated FW info.
	MapHTTPGetAggregation = "ftmap/v0/aggregation"

	// MapPrefixStrata is the number of prefix strata in the FT map.
	MapPrefixStrata = 1
	// MapTreeID is the unique tree ID salted into the map's hash functions.
	MapTreeID = 12345
)

// AggregatedFirmware represents the results of aggregating a single piece of firmware
// according to the rules described in #Aggregate().
type AggregatedFirmware struct {
	Index uint64
	Good  bool
}

// DeviceReleaseLog represents firmware releases found for a single device ID.
// Entries are ordered by their sequence in the original log.
type DeviceReleaseLog struct {
	DeviceID  string
	Revisions []uint64
}

// MapCheckpoint is a commitment to a map built from the FW Log at a given size.
// The map checkpoint contains the checkpoint of the log this was built from, with
// the number of entries consumed from that input log. This allows clients to check
// they are seeing the same version of the log as the map was built from. This also
// provides information to allow verifiers of the map to confirm correct construction.
type MapCheckpoint struct {
	// LogCheckpoint is the json encoded api.LogCheckpoint.
	LogCheckpoint []byte
	LogSize       uint64
	RootHash      []byte
	Revision      uint64
}

// MapTile is a subtree of the whole map.
type MapTile struct {
	// The path from the root of the map to the root of this tile.
	Path []byte
	// All non-empty leaves in this tile, sorted left-to-right.
	Leaves []MapTileLeaf
}

// MapTileLeaf is a leaf value of a MapTile.
// If it belongs to a leaf tile then this represents one of the values that the
// map commits to. Otherwise, this leaf represents the root of the subtree in
// the stratum below.
type MapTileLeaf struct {
	// The path from the root of the container MapTile to this leaf.
	Path []byte
	// The hash value being committed to.
	Hash []byte
}

// MapInclusionProof contains the value at the requested key and the proof to the
// requested Checkpoint.
type MapInclusionProof struct {
	Key   []byte
	Value []byte
	// Proof is all of the sibling hashes down the path, keyed by the bit length of the parent node ID.
	// A nil entry means that this branch is empty.
	// The parent node ID is used because the root does not have a sibling.
	Proof [][]byte
}

// String returns a compact printable representation of an InclusionProof.
func (l MapInclusionProof) String() string {
	return fmt.Sprintf("{key: 0x%x, value: 0x%x, proof: %x}", l.Key, l.Value, l.Proof)
}
