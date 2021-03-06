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

// Package client is a client for the serverless log.
package client

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/trillian-examples/formats/log"
	"github.com/google/trillian-examples/serverless/api"
	"github.com/google/trillian-examples/serverless/internal/layout"
	"github.com/google/trillian/merkle"
	"github.com/google/trillian/merkle/compact"
	"github.com/google/trillian/merkle/hashers"
	"github.com/google/trillian/merkle/logverifier"
)

// FetcherFunc is the signature of a function which can retrieve arbitrary files from
// a log's data storage, via whatever appropriate mechanism.
// The path parameter is relative to the root of the log storage.
type FetcherFunc func(path string) ([]byte, error)

// GetCheckpoint fetches and parses the latest LogState from the log.
func GetCheckpoint(f FetcherFunc) (*log.Checkpoint, error) {
	s, _, err := fetchCheckpointAndParse(f)
	return s, err
}

func fetchCheckpointAndParse(f FetcherFunc) (*log.Checkpoint, []byte, error) {
	cpRaw, err := f(layout.CheckpointPath)
	if err != nil {
		return nil, nil, err
	}
	cp := log.Checkpoint{}
	if _, err := (&cp).Unmarshal(cpRaw); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal checkpoint: %w", err)
	}
	return &cp, cpRaw, nil
}

// ProofBuilder knows how to build inclusion and consistency proofs from tiles.
// Since the tiles commit only to immutable nodes, the job of building proofs is slightly
// more complex as proofs can touch "ephemeral" nodes, so these need to be synthesized.
type ProofBuilder struct {
	cp        log.Checkpoint
	nodeCache NodeCache
	h         compact.HashFn
}

// NewProofBuilder creates a new ProofBuilder object for a given tree size.
// The returned ProofBuilder can be re-used for proofs related to a given tree size, but
// it is not thread-safe and should not be accessed concurrently.
func NewProofBuilder(cp log.Checkpoint, h compact.HashFn, f FetcherFunc) (*ProofBuilder, error) {
	pb := &ProofBuilder{
		cp:        cp,
		nodeCache: NewNodeCache(newTileFetcher(f)),
		h:         h,
	}

	hashes, err := FetchRangeNodes(cp.Size, &pb.nodeCache)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch range nodes: %w", err)
	}
	// Create a compact range which represents the state of the log.
	r, err := (&compact.RangeFactory{Hash: h}).NewRange(0, cp.Size, hashes)
	if err != nil {
		return nil, err
	}

	// Recreate the root hash so that:
	// a) we validate the self-integrity of the log state, and
	// b) we calculate (and cache) and ephemeral nodes present in the tree,
	//    this is important since they could be required by proofs.
	sr, err := r.GetRootHash(pb.nodeCache.SetEphemeralNode)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(cp.Hash, sr) {
		return nil, fmt.Errorf("invalid checkpoint hash %x, expected %x", cp.Hash, sr)
	}
	return pb, nil
}

// InclusionProof constructs an inclusion proof for the leaf at index in a tree of
// the given size.
// This function uses the passed-in function to retrieve tiles containing any log tree
// nodes necessary to build the proof.
func (pb *ProofBuilder) InclusionProof(index uint64) ([][]byte, error) {
	nodes, err := merkle.CalcInclusionProofNodeAddresses(int64(pb.cp.Size), int64(index), int64(pb.cp.Size))
	if err != nil {
		return nil, fmt.Errorf("failed to calculate inclusion proof node list: %w", err)
	}

	ret := make([][]byte, 0)
	// TODO(al) parallelise this.
	for _, n := range nodes {
		h, err := pb.nodeCache.GetNode(n.ID, pb.cp.Size)
		if err != nil {
			return nil, fmt.Errorf("failed to get node (%v): %w", n.ID, err)
		}
		ret = append(ret, h)
	}
	return ret, nil
}

// ConsistencyProof constructs a consistency proof between the two passed in tree sizes.
// This function uses the passed-in function to retrieve tiles containing any log tree
// nodes necessary to build the proof.
func (pb *ProofBuilder) ConsistencyProof(smaller, larger uint64) ([][]byte, error) {
	nodes, err := merkle.CalcConsistencyProofNodeAddresses(int64(smaller), int64(larger), int64(pb.cp.Size))
	if err != nil {
		return nil, fmt.Errorf("failed to calculate consistency proof node list: %w", err)
	}

	hashes := make([][]byte, 0)
	// TODO(al) parallelise this.
	for _, n := range nodes {
		h, err := pb.nodeCache.GetNode(n.ID, pb.cp.Size)
		if err != nil {
			return nil, fmt.Errorf("failed to get node (%v): %w", n.ID, err)
		}
		hashes = append(hashes, h)
	}
	return hashes, nil
}

// FetchRangeNodes returns the set of nodes representing the compact range covering
// a log of size s.
func FetchRangeNodes(s uint64, nc *NodeCache) ([][]byte, error) {
	nIDs := compact.RangeNodes(0, s)
	ret := make([][]byte, len(nIDs))
	for i, n := range nIDs {
		h, err := nc.GetNode(n, s)
		if err != nil {
			return nil, err
		}
		ret[i] = h
	}
	return ret, nil
}

// NodeCache hides the tiles abstraction away, and improves
// performance by caching tiles it's seen.
// Not threadsafe, and intended to be only used throughout the course
// of a single request.
type NodeCache struct {
	ephemeral map[compact.NodeID][]byte
	tiles     map[tileKey]api.Tile
	getTile   GetTileFunc
}

// GetTileFunc is the signature of a function which knows how to fetch a
// specific tile.
type GetTileFunc func(level, index, logSize uint64) (*api.Tile, error)

// tileKey is used as a key in nodeCache's tile map.
type tileKey struct {
	tileLevel uint64
	tileIndex uint64
}

// NewNodeCache creates a new nodeCache instance.
func NewNodeCache(f GetTileFunc) NodeCache {
	return NodeCache{
		ephemeral: make(map[compact.NodeID][]byte),
		tiles:     make(map[tileKey]api.Tile),
		getTile:   f,
	}
}

// SetEphemeralNode stored a derived "ephemeral" tree node.
func (n *NodeCache) SetEphemeralNode(id compact.NodeID, h []byte) {
	n.ephemeral[id] = h
}

// GetNode returns the internal log tree node hash for the specified node ID.
// A previously set ephemeral node will be returned if id matches, otherwise
// the tile containing the requested node will be fetched and cached, and the
// node hash returned.
func (n *NodeCache) GetNode(id compact.NodeID, logSize uint64) ([]byte, error) {
	// First check for ephemeral nodes:
	if e := n.ephemeral[id]; len(e) != 0 {
		return e, nil
	}
	// Otherwise look in fetched tiles:
	tileLevel, tileIndex, nodeLevel, nodeIndex := layout.NodeCoordsToTileAddress(uint64(id.Level), uint64(id.Index))
	tKey := tileKey{tileLevel, tileIndex}
	t, ok := n.tiles[tKey]
	if !ok {
		tile, err := n.getTile(tileLevel, tileIndex, logSize)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch tile: %w", err)
		}
		t = *tile
		n.tiles[tKey] = *tile
	}
	node := t.Nodes[api.TileNodeKey(nodeLevel, nodeIndex)]
	if node == nil {
		return nil, fmt.Errorf("node %v (tile coords [%d,%d]/[%d,%d]) unknown", id, tileLevel, tileIndex, nodeLevel, nodeIndex)
	}
	return node, nil
}

// newTileFetcher returns a GetTileFunc based on the passed in FetcherFunc.
func newTileFetcher(f FetcherFunc) GetTileFunc {
	return func(level, index, logSize uint64) (*api.Tile, error) {
		tileSize := layout.PartialTileSize(level, index, logSize)
		p := filepath.Join(layout.TilePath("", level, index, tileSize))
		t, err := f(p)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("failed to read tile at %q: %w", p, err)
			}
			return nil, err
		}

		var tile api.Tile
		if err := tile.UnmarshalText(t); err != nil {
			return nil, fmt.Errorf("failed to parse tile: %w", err)
		}
		return &tile, nil
	}
}

// LookupIndex fetches the leafhash->seq mapping file from the log, and returns
// its parsed contents.
func LookupIndex(f FetcherFunc, lh []byte) (uint64, error) {
	p := filepath.Join(layout.LeafPath("", lh))
	sRaw, err := f(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, fmt.Errorf("leafhash unknown (%w)", err)
		}
		return 0, fmt.Errorf("failed to fetch leafhash->seq file: %w", err)
	}
	return strconv.ParseUint(string(sRaw), 16, 64)
}

// LogStateTracker represents a client-side view of a target log's state.
// This tracker handles verification that updates to the tracked log state are
// consistent with previously seen states.
type LogStateTracker struct {
	Hasher   hashers.LogHasher
	Verifier logverifier.LogVerifier
	Fetcher  FetcherFunc

	// LatestConsistentRaw holds the raw bytes of the latest proven-consistent
	// LogState seen by this tracker.
	LatestConsistentRaw []byte

	// LatestConsistent is the deserialised form of LatestConsistentRaw
	LatestConsistent log.Checkpoint
}

// NewLogStateTracker creates a newly initialised tracker.
// If a serialised LogState representation is provided then this is used as the
// initial tracked state, otherwise a log state is fetched from the target log.
func NewLogStateTracker(f FetcherFunc, h hashers.LogHasher, checkpointRaw []byte) (LogStateTracker, error) {
	ret := LogStateTracker{
		Fetcher:          f,
		Hasher:           h,
		Verifier:         logverifier.New(h),
		LatestConsistent: log.Checkpoint{},
	}
	if len(checkpointRaw) > 0 {
		ret.LatestConsistentRaw = checkpointRaw
		if _, err := ret.LatestConsistent.Unmarshal(checkpointRaw); err != nil {
			return ret, err
		}
		return ret, nil
	}
	return ret, ret.Update()
}

// ErrInconsistency should be returned when there has been an error proving consistency
// between log states.
// The raw log state representations are included as-returned by the target log, this
// ensures that evidence of inconsistent log updates are available to the caller of
// the method(s) returning this error.
type ErrInconsistency struct {
	SmallerRaw []byte
	LargerRaw  []byte
	Proof      [][]byte

	Wrapped error
}

func (e ErrInconsistency) Unwrap() error {
	return e.Wrapped
}

func (e ErrInconsistency) Error() string {
	return fmt.Sprintf("log consistency check failed: %s", e.Wrapped)
}

// Update attempts to update the local view of the target log's state.
// If a more recent logstate is found, this method will attempt to prove
// that it is consistent with the local state before updating the tracker's
// view.
func (lst *LogStateTracker) Update() error {
	c, cRaw, err := fetchCheckpointAndParse(lst.Fetcher)
	if err != nil {
		return err
	}
	if lst.LatestConsistent.Size > 0 {
		if c.Size > lst.LatestConsistent.Size {
			builder, err := NewProofBuilder(*c, lst.Hasher.HashChildren, lst.Fetcher)
			if err != nil {
				return fmt.Errorf("failed to create proof builder: %w", err)
			}
			p, err := builder.ConsistencyProof(lst.LatestConsistent.Size, c.Size)
			if err != nil {
				return err
			}
			if err := lst.Verifier.VerifyConsistencyProof(int64(lst.LatestConsistent.Size), int64(c.Size), lst.LatestConsistent.Hash, c.Hash, p); err != nil {
				return ErrInconsistency{
					SmallerRaw: lst.LatestConsistentRaw,
					LargerRaw:  cRaw,
					Proof:      p,
					Wrapped:    err,
				}
			}
		}
	}
	lst.LatestConsistentRaw, lst.LatestConsistent = cRaw, *c
	return nil
}
