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
type ConfigNode struct {
	// DeploymentID is used to determine whether two NodeHost instances belong to
	// the same deployment and thus allowed to communicate with each other. This
	// helps to prvent accidentially misconfigured NodeHost instances to cause
	// data corruption errors by sending out of context messages to unrelated
	// Raft nodes.
	// For a particular dragonboat based application, you can set DeploymentID
	// to the same uint64 value on all production NodeHost instances, then use
	// different DeploymentID values on your staging and dev environment. It is
	// also recommended to use different DeploymentID values for different
	// dragonboat based applications.
	// When not set, the default value 0 will be used as the deployment ID and
	// thus allowing all NodeHost instances with deployment ID 0 to communicate
	// with each other.
	DeploymentID uint64 `mapstructure:"deployment_id" json:"deployment_id" yaml:"deployment_id" toml:"deployment_id"`

	// WALDir is the directory used for storing the WAL of Raft entries. It is
	// recommended to use low latency storage such as NVME SSD with power loss
	// protection to store such WAL data. Leave WALDir to have zero value will
	// have everything stored in NodeHostDir.
	WALDir string `mapstructure:"wal_dir" json:"wal_dir" yaml:"wal_dir" toml:"wal_dir" validate:"omitempty,printascii"`

	// NodeHostDir is where everything else is stored.
	NodeHostDir string `mapstructure:"node_host_dir" json:"node_host_dir" yaml:"node_host_dir" toml:"node_host_dir" validate:"omitempty,printascii"`

	//nolint #godox
	// RTTMillisecond defines the average Rround Trip Time (RTT) in milliseconds
	// between two NodeHost instances. Such a RTT interval is internally used as
	// a logical clock tick, Raft heartbeat and election intervals are both
	// defined in term of how many such logical clock ticks (RTT intervals).
	// Note that RTTMillisecond is the combined delays between two NodeHost
	// instances including all delays caused by network transmission, delays
	// caused by NodeHost queuing and processing. As an example, when fully
	// loaded, the average Rround Trip Time between two of our NodeHost instances
	// used for benchmarking purposes is up to 500 microseconds when the ping time
	// between them is 100 microseconds. Set RTTMillisecond to 1 when it is less
	// than 1 million in your environment.
	RTTMillisecond uint64 `mapstructure:"rtt_millisecond" json:"rtt_millisecond" yaml:"rtt_millisecond" toml:"rtt_millisecond"`

	// RaftAddress is a DNS name:port or IP:port address used by the transport
	// module for exchanging Raft messages, snapshots and metadata between
	// NodeHost instances. It should be set to the public address that can be
	// accessed from remote NodeHost instances.
	//
	// When the NodeHostConfig.ListenAddress field is empty, NodeHost listens on
	// RaftAddress for incoming Raft messages. When hostname or domain name is
	// used, it will be resolved to IPv4 addresses first and Dragonboat listens
	// to all resolved IPv4 addresses.
	//
	// By default, the RaftAddress value is not allowed to change between NodeHost
	// restarts. AddressByNodeHostID should be set to true when the RaftAddress
	// value might change after restart.
	RaftAddress string `mapstructure:"raft_address" json:"raft_address" yaml:"raft_address" toml:"raft_address" validate:"omitempty,printascii"`

	//nolint #godox
	// AddressByNodeHostID indicates that NodeHost instances should be addressed
	// by their NodeHostID values. This feature is usually used when only dynamic
	// addresses are available. When enabled, NodeHostID values should be used
	// as the target parameter when calling NodeHost's StartCluster,
	// RequestAddNode, RequestAddObserver and RequestAddWitness methods.
	//
	// Enabling AddressByNodeHostID also enables the internal gossip service,
	// NodeHostConfig.Gossip must be configured to control the behaviors of the
	// gossip service.
	//
	// Note that once enabled, the AddressByNodeHostID setting can not be later
	// disabled after restarts.
	//
	// Please see the godocs of the NodeHostConfig.Gossip field for a detailed
	// example on how AddressByNodeHostID and gossip works.
	AddressByNodeHostID bool `mapstructure:"address_by_node_host_id" json:"address_by_node_host_id" yaml:"address_by_node_host_id" toml:"address_by_node_host_id"`

	// ListenAddress is an optional field in the hostname:port or IP:port address
	// form used by the transport module to listen on for Raft message and
	// snapshots. When the ListenAddress field is not set, The transport module
	// listens on RaftAddress. If 0.0.0.0 is specified as the IP of the
	// ListenAddress, Dragonboat listens to the specified port on all network
	// interfaces. When hostname or domain name is used, it will be resolved to
	// IPv4 addresses first and Dragonboat listens to all resolved IPv4 addresses.
	ListenAddress string `mapstructure:"listen_address" json:"listen_address" yaml:"listen_address" toml:"listen_address" validate:"omitempty,hostname_port"`

	// MutualTLS defines whether to use mutual TLS for authenticating servers
	// and clients. Insecure communication is used when MutualTLS is set to
	// False.
	// See https://github.com/lni/dragonboat/wiki/TLS-in-Dragonboat for more
	// details on how to use Mutual TLS.
	MutualTLS bool `mapstructure:"mutual_tls" json:"mutual_tls" yaml:"tls" toml:"tls"`

	// CAFile is the path of the CA certificate file. This field is ignored when
	// MutualTLS is false.
	CAFile string `mapstructure:"ca_file" json:"ca_file" yaml:"ca_file" toml:"ca_file"`

	// CertFile is the path of the node certificate file. This field is ignored
	// when MutualTLS is false.
	CertFile string `mapstructure:"cert_file" json:"cert_file" yaml:"cert_file" toml:"cert_file"`

	// KeyFile is the path of the node key file. This field is ignored when
	// MutualTLS is false.
	KeyFile string `mapstructure:"key_file" json:"key_file" yaml:"key_file" toml:"key_file"`

	// EnableMetrics determines whether health metrics in Prometheus format should
	// be enabled.
	EnableMetrics bool `mapstructure:"enable_metrics" json:"enable_metrics" yaml:"enable_metrics" toml:"enable_metrics"`

	// MaxSendQueueSize is the maximum size in bytes of each send queue.
	// Once the maximum size is reached, further replication messages will be
	// dropped to restrict memory usage. When set to 0, it means the send queue
	// size is unlimited.
	MaxSendQueueSize uint64 `mapstructure:"max_send_queue_size" json:"max_send_queue_size" yaml:"max_send_queue_size" toml:"max_send_queue_size"`

	// MaxReceiveQueueSize is the maximum size in bytes of each receive queue.
	// Once the maximum size is reached, further replication messages will be
	// dropped to restrict memory usage. When set to 0, it means the queue size
	// is unlimited.
	MaxReceiveQueueSize uint64 `mapstructure:"max_receive_queue_size" json:"max_receive_queue_size" yaml:"max_receive_queue_size" toml:"max_receive_queue_size"`

	// MaxSnapshotSendBytesPerSecond defines how much snapshot data can be sent
	// every second for all Raft clusters managed by the NodeHost instance.
	// The default value 0 means there is no limit set for snapshot streaming.
	MaxSnapshotSendBytesPerSecond uint64 `mapstructure:"max_snapshot_send_bytes_per_second" json:"max_snapshot_send_bytes_per_second" yaml:"max_snapshot_send_bytes_per_second" toml:"max_snapshot_send_bytes_per_second"`

	// MaxSnapshotRecvBytesPerSecond defines how much snapshot data can be
	// received each second for all Raft clusters managed by the NodeHost instance.
	// The default value 0 means there is no limit for receiving snapshot data.
	MaxSnapshotRecvBytesPerSecond uint64 `mapstructure:"max_snapshot_recv_bytes_per_second" json:"max_snapshot_recv_bytes_per_second" yaml:"max_snapshot_recv_bytes_per_second" toml:"max_snapshot_recv_bytes_per_second"`

	// NotifyCommit specifies whether clients should be notified when their
	// regular proposals and config change requests are committed. By default,
	// commits are not notified, clients are only notified when their proposals
	// are both committed and applied.
	NotifyCommit bool `mapstructure:"notify_commit" json:"notify_commit" yaml:"notify_commit" toml:"notify_commit"`

	// Gossip contains configurations for the gossip service. When the
	// AddressByNodeHostID field is set to true, each NodeHost instance will use
	// an internal gossip service to exchange knowledges of known NodeHost
	// instances including their RaftAddress and NodeHostID values. This Gossip
	// field contains configurations that controls how the gossip service works.
	//
	// As an detailed example on how to use the gossip service in the situation
	// where all available machines have dynamically assigned IPs on reboot -
	//
	// Consider that there are three NodeHost instances on three machines, each
	// of them has a dynamically assigned IP address which will change on reboot.
	// NodeHostConfig.RaftAddress should be set to the current address that can be
	// reached by remote NodeHost instance. In this example, we will assume they
	// are
	//
	// 10.0.0.100:24000
	// 10.0.0.200:24000
	// 10.0.0.300:24000
	//
	// To use these machines, first enable the NodeHostConfig.AddressByNodeHostID
	// field and start the NodeHost instances. The NodeHostID value of each
	// NodeHost instance can be obtained by calling NodeHost.ID(). Let's say they
	// are
	//
	// "nhid-xxxxx",
	// "nhid-yyyyy",
	// "nhid-zzzzz".
	//
	// All these NodeHostID are fixed, they will never change after reboots.
	//
	// When starting Raft nodes or requesting new nodes to be added, use the above
	// mentioned NodeHostID values as the target parameters (which are of the
	// Target type). Let's say we want to start a Raft Node as a part of a three
	// replicas Raft cluster, the initialMembers parameter of the StartCluster
	// method can be set to
	//
	// initialMembers := map[uint64]Target {
	// 	 1: "nhid-xxxxx",
	//   2: "nhid-yyyyy",
	//   3: "nhid-zzzzz",
	// }
	//
	// This indicates that node 1 of the cluster will be running on the NodeHost
	// instance identified by the NodeHostID value "nhid-xxxxx", node 2 of the
	// same cluster will be running on the NodeHost instance identified by the
	// NodeHostID value of "nhid-yyyyy" and so on.
	//
	// The internal gossip service exchanges NodeHost details, including their
	// NodeHostID and RaftAddress values, with all other known NodeHost instances.
	// Thanks to the nature of gossip, it will eventually allow each NodeHost
	// instance to be aware of the current details of all NodeHost instances.
	// As a result, let's say when Raft node 1 wants to send a Raft message to
	// node 2, it first figures out that node 2 is running on the NodeHost
	// identified by the NodeHostID value "nhid-yyyyy", RaftAddress information
	// from the gossip service further shows that "nhid-yyyyy" maps to a machine
	// currently reachable at 10.0.0.200:24000. Raft messages can thus be
	// delivered.
	//
	// The Gossip field here is used to configure how the gossip service works.
	// In this example, let's say we choose to use the following configurations
	// for those three NodeHost instaces.
	//
	// GossipConfig {
	//   BindAddress: "10.0.0.100:24001",
	//   Seed: []string{10.0.0.200:24001},
	// }
	//
	// GossipConfig {
	//   BindAddress: "10.0.0.200:24001",
	//   Seed: []string{10.0.0.300:24001},
	// }
	//
	// GossipConfig {
	//   BindAddress: "10.0.0.300:24001",
	//   Seed: []string{10.0.0.100:24001},
	// }
	//
	// For those three machines, the gossip component listens on
	// "10.0.0.100:24001", "10.0.0.200:24001" and "10.0.0.300:24001" respectively
	// for incoming gossip messages. The Seed field is a list of known gossip end
	// points the local gossip service will try to talk to. The Seed field doesn't
	// need to include all gossip end points, a few well connected nodes in the
	// gossip network is enough.
	Gossip ConfigGossip `mapstructure:"gossip" json:"gossip" yaml:"gossip" toml:"gossip" validate:"omitempty,dive"`

	// Expert contains options for expert users who are familiar with the internals
	// of Dragonboat. Users are recommended not to use this field unless
	// absoloutely necessary. It is important to note that any change to this field
	// may cause an existing instance unable to restart, it may also cause negative
	// performance impacts.
	Expert ConfigExpert `mapstructure:"expert" json:"expert" yaml:"expert" toml:"expert" validate:"omitempty,dive"`
}

func (c ConfigNode) GetDGBConfigNodeHost() dgbcfg.NodeHostConfig {
	d := dgbcfg.NodeHostConfig{
		DeploymentID:  c.DeploymentID,
		WALDir:        c.WALDir,
		NodeHostDir:   c.NodeHostDir,
		RaftAddress:   c.RaftAddress,
		ListenAddress: c.ListenAddress,
		Gossip:        c.Gossip.GetDGBConfigGossip(),
		Expert:        c.Expert.GetDGBConfigExpert(),
	}

	if c.RTTMillisecond > 0 {
		d.RTTMillisecond = c.RTTMillisecond
	}

	if c.AddressByNodeHostID {
		d.AddressByNodeHostID = true
	}

	if c.MutualTLS && c.CAFile != "" && c.CertFile != "" && c.KeyFile != "" {
		d.MutualTLS = true
		d.CAFile = c.CAFile
		d.CertFile = c.CertFile
		d.KeyFile = c.KeyFile
	}

	if c.EnableMetrics {
		d.EnableMetrics = true
	}

	if c.MaxSendQueueSize > 0 {
		d.MaxSendQueueSize = c.MaxSendQueueSize
	}

	if c.MaxReceiveQueueSize > 0 {
		d.MaxReceiveQueueSize = c.MaxReceiveQueueSize
	}

	if c.MaxSnapshotSendBytesPerSecond > 0 {
		d.MaxSnapshotSendBytesPerSecond = c.MaxSnapshotSendBytesPerSecond
	}

	if c.MaxSnapshotRecvBytesPerSecond > 0 {
		d.MaxSnapshotRecvBytesPerSecond = c.MaxSnapshotRecvBytesPerSecond
	}

	if c.NotifyCommit {
		d.NotifyCommit = true
	}

	return d
}

func (c ConfigNode) Validate() liberr.Error {
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
