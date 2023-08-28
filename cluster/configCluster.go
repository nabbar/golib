//go:build amd64 || arm64 || arm64be || ppc64 || ppc64le || mips64 || mips64le || riscv64 || s390x || sparc64 || wasm
// +build amd64 arm64 arm64be ppc64 ppc64le mips64 mips64le riscv64 s390x sparc64 wasm

/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package cluster

import (
	"fmt"

	libval "github.com/go-playground/validator/v10"
	dgbcfg "github.com/lni/dragonboat/v3/config"
	liberr "github.com/nabbar/golib/errors"
)

// nolint #maligned
type ConfigCluster struct {
	// NodeID is a non-zero value used to identify a node within a Raft cluster.
	NodeID uint64 `mapstructure:"node_id" json:"node_id" yaml:"node_id" toml:"node_id"`

	// ClusterID is the unique value used to identify a Raft cluster.
	ClusterID uint64 `mapstructure:"cluster_id" json:"cluster_id" yaml:"cluster_id" toml:"cluster_id"`

	// CheckQuorum specifies whether the leader node should periodically check
	// non-leader node status and step down to become a follower node when it no
	// longer has the quorum.
	CheckQuorum bool `mapstructure:"check_quorum" json:"check_quorum" yaml:"check_quorum" toml:"check_quorum"`

	// ElectionRTT is the minimum number of message RTT between elections. Message
	// RTT is defined by NodeHostConfig.RTTMillisecond. The Raft paper suggests it
	// to be a magnitude greater than HeartbeatRTT, which is the interval between
	// two heartbeats. In Raft, the actual interval between elections is
	// randomized to be between ElectionRTT and 2 * ElectionRTT.
	//
	// As an example, assuming NodeHostConfig.RTTMillisecond is 100 millisecond,
	// to set the election interval to be 1 second, then ElectionRTT should be set
	// to 10.
	//
	// When CheckQuorum is enabled, ElectionRTT also defines the interval for
	// checking leader quorum.
	ElectionRTT uint64 `mapstructure:"election_rtt" json:"election_rtt" yaml:"election_rtt" toml:"election_rtt"`

	// HeartbeatRTT is the number of message RTT between heartbeats. Message
	// RTT is defined by NodeHostConfig.RTTMillisecond. The Raft paper suggest the
	// heartbeat interval to be close to the average RTT between nodes.
	//
	// As an example, assuming NodeHostConfig.RTTMillisecond is 100 millisecond,
	// to set the heartbeat interval to be every 200 milliseconds, then
	// HeartbeatRTT should be set to 2.
	HeartbeatRTT uint64 `mapstructure:"heartbeat_rtt" json:"heartbeat_rtt" yaml:"heartbeat_rtt" toml:"heartbeat_rtt"`

	// SnapshotEntries defines how often the state machine should be snapshotted
	// automcatically. It is defined in terms of the number of applied Raft log
	// entries. SnapshotEntries can be set to 0 to disable such automatic
	// snapshotting.
	//
	// When SnapshotEntries is set to N, it means a snapshot is created for
	// roughly every N applied Raft log entries (proposals). This also implies
	// that sending N log entries to a follower is more expensive than sending a
	// snapshot.
	//
	// Once a snapshot is generated, Raft log entries covered by the new snapshot
	// can be compacted. This involves two steps, redundant log entries are first
	// marked as deleted, then they are physically removed from the underlying
	// storage when a LogDB compaction is issued at a later stage. See the godoc
	// on CompactionOverhead for details on what log entries are actually removed
	// and compacted after generating a snapshot.
	//
	// Once automatic snapshotting is disabled by setting the SnapshotEntries
	// field to 0, users can still use NodeHost's RequestSnapshot or
	// SyncRequestSnapshot methods to manually request snapshots.
	SnapshotEntries uint64 `mapstructure:"snapshot_entries" json:"snapshot_entries" yaml:"snapshot_entries" toml:"snapshot_entries"`

	// CompactionOverhead defines the number of most recent entries to keep after
	// each Raft log compaction. Raft log compaction is performance automatically
	// every time when a snapshot is created.
	//
	// For example, when a snapshot is created at let's say index 10,000, then all
	// Raft log entries with index <= 10,000 can be removed from that node as they
	// have already been covered by the created snapshot image. This frees up the
	// maximum storage space but comes at the cost that the full snapshot will
	// have to be sent to the follower if the follower requires any Raft log entry
	// at index <= 10,000. When CompactionOverhead is set to say 500, Dragonboat
	// then compacts the Raft log up to index 9,500 and keeps Raft log entries
	// between index (9,500, 1,0000]. As a result, the node can still replicate
	// Raft log entries between index (9,500, 1,0000] to other peers and only fall
	// back to stream the full snapshot if any Raft log entry with index <= 9,500
	// is required to be replicated.
	CompactionOverhead uint64 `mapstructure:"compaction_overhead" json:"compaction_overhead" yaml:"compaction_overhead" toml:"compaction_overhead"`

	// OrderedConfigChange determines whether Raft membership change is enforced
	// with ordered config change ID.
	OrderedConfigChange bool `mapstructure:"ordered_config_change" json:"ordered_config_change" yaml:"ordered_config_change" toml:"ordered_config_change"`

	// MaxInMemLogSize is the target size in bytes allowed for storing in memory
	// Raft logs on each Raft node. In memory Raft logs are the ones that have
	// not been applied yet.
	// MaxInMemLogSize is a target value implemented to prevent unbounded memory
	// growth, it is not for precisely limiting the exact memory usage.
	// When MaxInMemLogSize is 0, the target is set to math.MaxUint64. When
	// MaxInMemLogSize is set and the target is reached, error will be returned
	// when clients try to make new proposals.
	// MaxInMemLogSize is recommended to be significantly larger than the biggest
	// proposal you are going to use.
	MaxInMemLogSize uint64 `mapstructure:"max_in_mem_log_size" json:"max_in_mem_log_size" yaml:"max_in_mem_log_size" toml:"max_in_mem_log_size"`

	// SnapshotCompressionType is the compression type to use for compressing
	// generated snapshot data. No compression is used by default.
	SnapshotCompressionType dgbcfg.CompressionType `mapstructure:"snapshot_compression" json:"snapshot_compression" yaml:"snapshot_compression" toml:"snapshot_compression"`

	// EntryCompressionType is the compression type to use for compressing the
	// payload of user proposals. When Snappy is used, the maximum proposal
	// payload allowed is roughly limited to 3.42GBytes. No compression is used
	// by default.
	EntryCompressionType dgbcfg.CompressionType `mapstructure:"entry_compression" json:"entry_compression" yaml:"entry_compression" toml:"entry_compression"`

	// DisableAutoCompactions disables auto compaction used for reclaiming Raft
	// log entry storage spaces. By default, compaction request is issued every
	// time when a snapshot is created, this helps to reclaim disk spaces as
	// soon as possible at the cost of immediate higher IO overhead. Users can
	// disable such auto compactions and use NodeHost.RequestCompaction to
	// manually request such compactions when necessary.
	DisableAutoCompactions bool `mapstructure:"disable_auto_compactions" json:"disable_auto_compactions" yaml:"disable_auto_compactions" toml:"disable_auto_compactions"`

	// IsObserver indicates whether this is an observer Raft node without voting
	// power. Described as non-voting members in the section 4.2.1 of Diego
	// Ongaro's thesis, observer nodes are usually used to allow a new node to
	// join the cluster and catch up with other existing ndoes without impacting
	// the availability. Extra observer nodes can also be introduced to serve
	// read-only requests without affecting system write throughput.
	//
	// Observer support is currently experimental.
	IsObserver bool `mapstructure:"is_observer" json:"is_observer" yaml:"is_observer" toml:"is_observer"`

	// IsWitness indicates whether this is a witness Raft node without actual log
	// replication and do not have state machine. It is mentioned in the section
	// 11.7.2 of Diego Ongaro's thesis.
	//
	// Witness support is currently experimental.
	IsWitness bool `mapstructure:"is_witness" json:"is_witness" yaml:"is_witness" toml:"is_witness"`

	// Quiesce specifies whether to let the Raft cluster enter quiesce mode when
	// there is no cluster activity. Clusters in quiesce mode do not exchange
	// heartbeat messages to minimize bandwidth consumption.
	//
	// Quiesce support is currently experimental.
	Quiesce bool `mapstructure:"quiesce" json:"quiesce" yaml:"quiesce" toml:"quiesce"`
}

func (c ConfigCluster) GetDGBConfigCluster() dgbcfg.Config {
	d := dgbcfg.Config{
		NodeID:                  c.NodeID,
		ClusterID:               c.ClusterID,
		SnapshotCompressionType: 0,
		EntryCompressionType:    0,
	}

	if c.CheckQuorum {
		d.CheckQuorum = true
	}

	if c.ElectionRTT > 0 {
		d.ElectionRTT = c.ElectionRTT
	}

	if c.HeartbeatRTT > 0 {
		d.HeartbeatRTT = c.HeartbeatRTT
	}

	if c.SnapshotEntries > 0 {
		d.SnapshotEntries = c.SnapshotEntries
	}

	if c.CompactionOverhead > 0 {
		d.CompactionOverhead = c.CompactionOverhead
	}

	if c.OrderedConfigChange {
		d.OrderedConfigChange = true
	}

	if c.MaxInMemLogSize > 0 {
		d.MaxInMemLogSize = c.MaxInMemLogSize
	}

	if c.DisableAutoCompactions {
		d.DisableAutoCompactions = true
	}

	if c.IsObserver {
		d.IsObserver = true
	}

	if c.IsWitness {
		d.IsWitness = true
	}

	if c.Quiesce {
		d.Quiesce = true
	}

	//nolint #exhaustive
	switch c.SnapshotCompressionType {
	case dgbcfg.Snappy:
		d.SnapshotCompressionType = dgbcfg.Snappy
	default:
		d.SnapshotCompressionType = dgbcfg.NoCompression
	}

	//nolint #exhaustive
	switch c.EntryCompressionType {
	case dgbcfg.Snappy:
		d.EntryCompressionType = dgbcfg.Snappy
	default:
		d.EntryCompressionType = dgbcfg.NoCompression
	}

	return d
}

func (c ConfigCluster) Validate() liberr.Error {
	err := ErrorValidateConfig.Error(nil)

	if er := libval.New().Struct(c); er != nil {
		if e, ok := er.(*libval.InvalidValidationError); ok {
			err.Add(e)
		}

		for _, e := range er.(libval.ValidationErrors) {
			//nolint goerr113
			err.Add(fmt.Errorf("config field '%s' is not validated by constraint '%s'", e.Namespace(), e.ActualTag()))
		}
	}

	if err.HasParent() {
		return err
	}

	return nil
}
