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
	"context"
	"time"

	dgbclt "github.com/lni/dragonboat/v3"
	"github.com/lni/dragonboat/v3/client"
	dgbcfg "github.com/lni/dragonboat/v3/config"
	sm "github.com/lni/dragonboat/v3/statemachine"
)

type Cluster interface {
	// NodeHostConfig returns the NodeHostConfig instance used for configuring this
	// NodeHost instance.
	NodeHostConfig() dgbcfg.NodeHostConfig

	// RaftAddress returns the Raft address of the NodeHost instance, it is the
	// network address by which the NodeHost can be reached by other NodeHost
	// instances for exchanging Raft messages, snapshots and other metadata.
	RaftAddress() string

	// ID returns the string representation of the NodeHost ID value. The NodeHost
	// ID is assigned to each NodeHost on its initial creation and it can be used
	// to uniquely identify the NodeHost instance for its entire life cycle. When
	// the system is running in the AddressByNodeHost mode, it is used as the target
	// value when calling the StartCluster, RequestAddNode, RequestAddObserver,
	// RequestAddWitness methods.
	ID() string

	// StartCluster adds the specified Raft cluster node to the NodeHost and starts
	// the node to make it ready for accepting incoming requests. The node to be
	// started is backed by a regular state machine that implements the
	// sm.IStateMachine interface.
	//
	// The input parameter initialMembers is a map of node ID to node target for all
	// Raft cluster's initial member nodes. By default, the target is the
	// RaftAddress value of the NodeHost where the node will be running. When running
	// in the AddressByNodeHostID mode, target should be set to the NodeHostID value
	// of the NodeHost where the node will be running. See the godoc of NodeHost's ID
	// method for the full definition of NodeHostID. For the same Raft cluster, the
	// same initialMembers map should be specified when starting its initial member
	// nodes on distributed NodeHost instances.
	//
	// The join flag indicates whether the node is a new node joining an existing
	// cluster. create is a factory function for creating the IStateMachine instance,
	// cfg is the configuration instance that will be passed to the underlying Raft
	// node object, the cluster ID and node ID of the involved node are specified in
	// the ClusterID and NodeID fields of the provided cfg parameter.
	//
	// Note that this method is not for changing the membership of the specified
	// Raft cluster, it launches a node that is already a member of the Raft
	// cluster.
	//
	// As a summary, when -
	//  - starting a brand new Raft cluster, set join to false and specify all initial
	//    member node details in the initialMembers map.
	//  - joining a new node to an existing Raft cluster, set join to true and leave
	//    the initialMembers map empty. This requires the joining node to have already
	//    been added as a member node of the Raft cluster.
	//  - restarting an crashed or stopped node, set join to false and leave the
	//    initialMembers map to be empty. This applies to both initial member nodes
	//    and those joined later.
	StartCluster(initialMembers map[uint64]dgbclt.Target, join bool, create sm.CreateStateMachineFunc, cfg dgbcfg.Config) error

	// StartConcurrentCluster is similar to the StartCluster method but it is used
	// to start a Raft node backed by a concurrent state machine.
	StartConcurrentCluster(initialMembers map[uint64]dgbclt.Target, join bool, create sm.CreateConcurrentStateMachineFunc, cfg dgbcfg.Config) error

	// StartOnDiskCluster is similar to the StartCluster method but it is used to
	// start a Raft node backed by an IOnDiskStateMachine.
	StartOnDiskCluster(initialMembers map[uint64]dgbclt.Target, join bool, create sm.CreateOnDiskStateMachineFunc, cfg dgbcfg.Config) error

	// Stop stops all Raft nodes managed by the NodeHost instance, it also closes
	// all internal components such as the transport and LogDB modules.
	Stop()

	// StopCluster removes and stops the Raft node associated with the specified
	// Raft cluster from the NodeHost. The node to be removed and stopped is
	// identified by the clusterID value.
	//
	// Note that this is not the membership change operation to remove the node
	// from the Raft cluster.
	StopCluster(clusterID uint64) error

	// StopNode removes the specified Raft cluster node from the NodeHost and
	// stops that running Raft node.
	//
	// Note that this is not the membership change operation to remove the node
	// from the Raft cluster.
	StopNode(clusterID uint64, nodeID uint64) error

	// SyncPropose makes a synchronous proposal on the Raft cluster specified by
	// the input client session object. The specified context parameter must has
	// the timeout value set.
	//
	// SyncPropose returns the result returned by IStateMachine or
	// IOnDiskStateMachine's Update method, or the error encountered. The input
	// byte slice can be reused for other purposes immediate after the return of
	// this method.
	//
	// After calling SyncPropose, unless NO-OP client session is used, it is
	// caller's responsibility to update the client session instance accordingly
	// based on SyncPropose's outcome. Basically, when a ErrTimeout error is
	// returned, application can retry the same proposal without updating the
	// client session instance. When ErrInvalidSession error is returned, it
	// usually means the session instance has been evicted from the server side,
	// the Raft paper recommends to crash the client in this highly unlikely
	// event. When the proposal completed successfully, caller must call
	// client.ProposalCompleted() to get it ready to be used in future proposals.
	SyncPropose(ctx context.Context, session *client.Session, cmd []byte) (sm.Result, error)

	// SyncRead performs a synchronous linearizable read on the specified Raft
	// cluster. The specified context parameter must has the timeout value set. The
	// query byte slice specifies what to query, it will be passed to the Lookup
	// method of the IStateMachine or IOnDiskStateMachine after the system
	// determines that it is safe to perform the local read on IStateMachine or
	// IOnDiskStateMachine. It returns the query result from the Lookup method or
	// the error encountered.
	SyncRead(ctx context.Context, clusterID uint64, query interface{}) (interface{}, error)

	// SyncGetClusterMembership is a rsynchronous method that queries the membership
	// information from the specified Raft cluster. The specified context parameter
	// must has the timeout value set.
	//
	// SyncGetClusterMembership guarantees that the returned membership information
	// is linearizable.
	SyncGetClusterMembership(ctx context.Context, clusterID uint64) (*dgbclt.Membership, error)

	// GetClusterMembership returns the membership information from the specified
	// Raft cluster. The specified context parameter must has the timeout value
	// set.
	//
	// GetClusterMembership guarantees that the returned membership information is
	// linearizable. This is a synchronous method meaning it will only return after
	// its confirmed completion, failure or timeout.
	//
	// Deprecated: Use NodeHost.SyncGetClusterMembership instead.
	// NodeHost.GetClusterMembership will be removed in v4.0.
	GetClusterMembership(ctx context.Context, clusterID uint64) (*dgbclt.Membership, error)

	// GetLeaderID returns the leader node ID of the specified Raft cluster based
	// on local node's knowledge. The returned boolean value indicates whether the
	// leader information is available.
	GetLeaderID(clusterID uint64) (uint64, bool, error)

	// GetNoOPSession returns a NO-OP client session ready to be used for making
	// proposals. The NO-OP client session is a dummy client session that will not
	// be checked or enforced. Use this No-OP client session when you want to ignore
	// features provided by client sessions. A NO-OP client session is not
	// registered on the server side and thus not required to be closed at the end
	// of its life cycle.
	//
	// Returned NO-OP client session instance can be concurrently used in multiple
	// goroutines.
	//
	// Use this NO-OP client session when your IStateMachine provides idempotence in
	// its own implementation.
	//
	// NO-OP client session must be used for making proposals on IOnDiskStateMachine
	// based state machine.
	GetNoOPSession(clusterID uint64) *client.Session

	// GetNewSession starts a synchronous proposal to create, register and return
	// a new client session object for the specified Raft cluster. The specified
	// context parameter must has the timeout value set.
	//
	// A client session object is used to ensure that a retried proposal, e.g.
	// proposal retried after timeout, will not be applied more than once into the
	// IStateMachine.
	//
	// Returned client session instance should not be used concurrently. Use
	// multiple client sessions when making concurrent proposals.
	//
	// Deprecated: Use NodeHost.SyncGetSession instead. NodeHost.GetNewSession will
	// be removed in v4.0.
	GetNewSession(ctx context.Context, clusterID uint64) (*client.Session, error)

	// CloseSession closes the specified client session by unregistering it
	// from the system. The specified context parameter must has the timeout value
	// set. This is a synchronous method meaning it will only return after its
	// confirmed completion, failure or timeout.
	//
	// Closed client session should no longer be used in future proposals.
	//
	// Deprecated: Use NodeHost.SyncCloseSession instead. NodeHost.CloseSession will
	// be removed in v4.0.
	CloseSession(ctx context.Context, session *client.Session) error

	// SyncGetSession starts a synchronous proposal to create, register and return
	// a new client session object for the specified Raft cluster. The specified
	// context parameter must has the timeout value set.
	//
	// A client session object is used to ensure that a retried proposal, e.g.
	// proposal retried after timeout, will not be applied more than once into the
	// state machine.
	//
	// Returned client session instance should not be used concurrently. Use
	// multiple client sessions when you need to concurrently start multiple
	// proposals.
	//
	// Client session is not supported by IOnDiskStateMachine based state machine.
	// NO-OP client session must be used for making proposals on IOnDiskStateMachine
	// based state machine.
	SyncGetSession(ctx context.Context, clusterID uint64) (*client.Session, error)

	// SyncCloseSession closes the specified client session by unregistering it
	// from the system. The specified context parameter must has the timeout value
	// set. This is a synchronous method meaning it will only return after its
	// confirmed completion, failure or timeout.
	//
	// Closed client session should no longer be used in future proposals.
	SyncCloseSession(ctx context.Context, cs *client.Session) error

	// Propose starts an asynchronous proposal on the Raft cluster specified by the
	// Session object. The input byte slice can be reused for other purposes
	// immediate after the return of this method.
	//
	// This method returns a RequestState instance or an error immediately.
	// Application can wait on the ResultC() channel of the returned RequestState
	// instance to get notified for the outcome of the proposal and access to the
	// result of the proposal.
	//
	// After the proposal is completed, i.e. RequestResult is received from the
	// ResultC() channel of the returned RequestState, unless NO-OP client session
	// is used, it is caller's responsibility to update the Session instance
	// accordingly based on the RequestResult.Code value. Basically, when
	// RequestTimeout is returned, you can retry the same proposal without updating
	// your client session instance, when a RequestRejected value is returned, it
	// usually means the session instance has been evicted from the server side,
	// the Raft paper recommends you to crash your client in this highly unlikely
	// event. When the proposal completed successfully with a RequestCompleted
	// value, application must call client.ProposalCompleted() to get the client
	// session ready to be used in future proposals.
	Propose(session *client.Session, cmd []byte, timeout time.Duration) (*dgbclt.RequestState, error)

	// ProposeSession starts an asynchronous proposal on the specified cluster
	// for client session related operations. Depending on the state of the specified
	// client session object, the supported operations are for registering or
	// unregistering a client session. Application can select on the ResultC()
	// channel of the returned RequestState instance to get notified for the
	// completion (RequestResult.Completed() is true) of the operation.
	ProposeSession(session *client.Session, timeout time.Duration) (*dgbclt.RequestState, error)

	// ReadIndex starts the asynchronous ReadIndex protocol used for linearizable
	// read on the specified cluster. This method returns a RequestState instance
	// or an error immediately. Application should wait on the ResultC() channel
	// of the returned RequestState object to get notified on the outcome of the
	// ReadIndex operation. On a successful completion, the ReadLocal method can
	// then be invoked to query the state of the IStateMachine or
	// IOnDiskStateMachine to complete the read operation with linearizability
	// guarantee.
	ReadIndex(clusterID uint64, timeout time.Duration) (*dgbclt.RequestState, error)

	// ReadLocalNode queries the Raft node identified by the input RequestState
	// instance. To ensure the IO linearizability, ReadLocalNode should only be
	// called after receiving a RequestCompleted notification from the ReadIndex
	// method. See ReadIndex's example for more details.
	ReadLocalNode(rs *dgbclt.RequestState, query interface{}) (interface{}, error)

	// NAReadLocalNode is a variant of ReadLocalNode, it uses byte slice as its
	// input and output data for read only queries to minimize extra heap
	// allocations caused by using interface{}. Users are recommended to use
	// ReadLocalNode unless performance is the top priority.
	//
	// As an optional method, the underlying state machine must implement the
	// statemachine.IExtended interface. NAReadLocalNode returns
	// statemachine.ErrNotImplemented if the underlying state machine does not
	// implement the statemachine.IExtended interface.
	NAReadLocalNode(rs *dgbclt.RequestState, query []byte) ([]byte, error)

	// StaleRead queries the specified Raft node directly without any
	// linearizability guarantee.
	//
	// Users are recommended to use the SyncRead method or a combination of the
	// ReadIndex and ReadLocalNode method to achieve linearizable read.
	StaleRead(clusterID uint64, query interface{}) (interface{}, error)

	// SyncRequestSnapshot is the synchronous variant of the RequestSnapshot
	// method. See RequestSnapshot for more details.
	//
	// The input ctx must has deadline set.
	//
	// SyncRequestSnapshot returns the index of the created snapshot or the error
	// encountered.
	SyncRequestSnapshot(ctx context.Context, clusterID uint64, opt dgbclt.SnapshotOption) (uint64, error)

	// RequestSnapshot requests a snapshot to be created asynchronously for the
	// specified cluster node. For each node, only one ongoing snapshot operation
	// is allowed.
	//
	// Users can use an option parameter to specify details of the requested
	// snapshot. For example, when the input SnapshotOption's Exported field is
	// True, a snapshot will be exported to the directory pointed by the ExportPath
	// field of the SnapshotOption instance. Such an exported snapshot is not
	// managed by the system and it is mainly used to repair the cluster when it
	// permanently loses its majority quorum. See the ImportSnapshot method in the
	// tools package for more details.
	//
	// When the Exported field of the input SnapshotOption instance is set to false,
	// snapshots created as the result of RequestSnapshot are managed by Dragonboat.
	// Users are not suppose to move, copy, modify or delete the generated snapshot.
	// Such requested snapshot will also trigger Raft log and snapshot compactions
	// similar to automatic snapshotting. Users need to subsequently call
	// RequestCompaction(), which can be far more I/O intensive, at suitable time to
	// actually reclaim disk spaces used by Raft log entries and snapshot metadata
	// records.
	//
	// When a snapshot is requested on a node backed by an IOnDiskStateMachine, only
	// the metadata portion of the state machine will be captured and saved.
	// Requesting snapshots on IOnDiskStateMachine based nodes are typically used to
	// trigger Raft log and snapshot compactions.
	//
	// RequestSnapshot returns a RequestState instance or an error immediately.
	// Applications can wait on the ResultC() channel of the returned RequestState
	// instance to get notified for the outcome of the create snasphot operation.
	// The RequestResult instance returned by the ResultC() channel tells the
	// outcome of the snapshot operation, when successful, the SnapshotIndex method
	// of the returned RequestResult instance reports the index of the created
	// snapshot.
	//
	// Requested snapshot operation will be rejected if there is already an existing
	// snapshot in the system at the same Raft log index.
	RequestSnapshot(clusterID uint64, opt dgbclt.SnapshotOption, timeout time.Duration) (*dgbclt.RequestState, error)

	// RequestCompaction requests a compaction operation to be asynchronously
	// executed in the background to reclaim disk spaces used by Raft Log entries
	// that have already been marked as removed. This includes Raft Log entries
	// that have already been included in created snapshots and Raft Log entries
	// that belong to nodes already permanently removed via NodeHost.RemoveData().
	//
	// By default, compaction is automatically issued after each snapshot is
	// captured. RequestCompaction can be used to manually trigger such compaction
	// when auto compaction is disabled by the DisableAutoCompactions option in
	// config.Config.
	//
	// The returned *SysOpState instance can be used to get notified when the
	// requested compaction is completed. ErrRejected is returned when there is
	// nothing to be reclaimed.
	RequestCompaction(clusterID uint64, nodeID uint64) (*dgbclt.SysOpState, error)

	// SyncRequestDeleteNode is the synchronous variant of the RequestDeleteNode
	// method. See RequestDeleteNode for more details.
	//
	// The input ctx must have its deadline set.
	SyncRequestDeleteNode(ctx context.Context, clusterID uint64, nodeID uint64, configChangeIndex uint64) error

	// SyncRequestAddNode is the synchronous variant of the RequestAddNode method.
	// See RequestAddNode for more details.
	//
	// The input ctx must have its deadline set.
	SyncRequestAddNode(ctx context.Context, clusterID uint64, nodeID uint64, target string, configChangeIndex uint64) error

	// SyncRequestAddObserver is the synchronous variant of the RequestAddObserver
	// method. See RequestAddObserver for more details.
	//
	// The input ctx must have its deadline set.
	SyncRequestAddObserver(ctx context.Context, clusterID uint64, nodeID uint64, target string, configChangeIndex uint64) error

	// SyncRequestAddWitness is the synchronous variant of the RequestAddWitness
	// method. See RequestAddWitness for more details.
	//
	// The input ctx must have its deadline set.
	SyncRequestAddWitness(ctx context.Context, clusterID uint64, nodeID uint64, target string, configChangeIndex uint64) error

	// RequestDeleteNode is a Raft cluster membership change method for requesting
	// the specified node to be removed from the specified Raft cluster. It starts
	// an asynchronous request to remove the node from the Raft cluster membership
	// list. Application can wait on the ResultC() channel of the returned
	// RequestState instance to get notified for the outcome.
	//
	// It is not guaranteed that deleted node will automatically close itself and
	// be removed from its managing NodeHost instance. It is application's
	// responsibility to call RemoveCluster on the right NodeHost instance to
	// actually have the cluster node removed from its managing NodeHost instance.
	//
	// Once a node is successfully deleted from a Raft cluster, it will not be
	// allowed to be added back to the cluster with the same node identity.
	//
	// When the raft cluster is created with the OrderedConfigChange config flag
	// set as false, the configChangeIndex parameter is ignored. Otherwise, it
	// should be set to the most recent Config Change Index value returned by the
	// SyncGetClusterMembership method. The requested delete node operation will be
	// rejected if other membership change has been applied since that earlier call
	// to the SyncGetClusterMembership method.
	RequestDeleteNode(clusterID uint64, nodeID uint64, configChangeIndex uint64, timeout time.Duration) (*dgbclt.RequestState, error)

	// RequestAddNode is a Raft cluster membership change method for requesting the
	// specified node to be added to the specified Raft cluster. It starts an
	// asynchronous request to add the node to the Raft cluster membership list.
	// Application can wait on the ResultC() channel of the returned RequestState
	// instance to get notified for the outcome.
	//
	// If there is already an observer with the same nodeID in the cluster, it will
	// be promoted to a regular node with voting power. The target parameter of the
	// RequestAddNode call is ignored when promoting an observer to a regular node.
	//
	// After the node is successfully added to the Raft cluster, it is application's
	// responsibility to call StartCluster on the target NodeHost instance to
	// actually start the Raft cluster node.
	//
	// Requesting a removed node back to the Raft cluster will always be rejected.
	//
	// By default, the target parameter is the RaftAddress of the NodeHost instance
	// where the new Raft node will be running. Note that fixed IP or static DNS
	// name should be used in RaftAddress in such default mode. When running in the
	// AddressByNodeHostID mode, target should be set to NodeHost's ID value which
	// can be obtained by calling the ID() method.
	//
	// When the Raft cluster is created with the OrderedConfigChange config flag
	// set as false, the configChangeIndex parameter is ignored. Otherwise, it
	// should be set to the most recent Config Change Index value returned by the
	// SyncGetClusterMembership method. The requested add node operation will be
	// rejected if other membership change has been applied since that earlier call
	// to the SyncGetClusterMembership method.
	RequestAddNode(clusterID uint64, nodeID uint64, target dgbclt.Target, configChangeIndex uint64, timeout time.Duration) (*dgbclt.RequestState, error)

	// RequestAddObserver is a Raft cluster membership change method for requesting
	// the specified node to be added to the specified Raft cluster as an observer
	// without voting power. It starts an asynchronous request to add the specified
	// node as an observer.
	//
	// Such observer is able to receive replicated states from the leader node, but
	// it is neither allowed to vote for leader, nor considered as a part of the
	// quorum when replicating state. An observer can be promoted to a regular node
	// with voting power by making a RequestAddNode call using its clusterID and
	// nodeID values. An observer can be removed from the cluster by calling
	// RequestDeleteNode with its clusterID and nodeID values.
	//
	// Application should later call StartCluster with config.Config.IsObserver
	// set to true on the right NodeHost to actually start the observer instance.
	//
	// By default, the target parameter is the RaftAddress of the NodeHost instance
	// where the new Raft node will be running. Note that fixed IP or static DNS
	// name should be used in RaftAddress in such default mode. When running in the
	// AddressByNodeHostID mode, target should be set to NodeHost's ID value which
	// can be obtained by calling the ID() method.
	//
	// When the Raft cluster is created with the OrderedConfigChange config flag
	// set as false, the configChangeIndex parameter is ignored. Otherwise, it
	// should be set to the most recent Config Change Index value returned by the
	// SyncGetClusterMembership method. The requested add observer operation will be
	// rejected if other membership change has been applied since that earlier call
	// to the SyncGetClusterMembership method.
	RequestAddObserver(clusterID uint64, nodeID uint64, target dgbclt.Target, configChangeIndex uint64, timeout time.Duration) (*dgbclt.RequestState, error)

	// RequestAddWitness is a Raft cluster membership change method for requesting
	// the specified node to be added as a witness to the given Raft cluster. It
	// starts an asynchronous request to add the specified node as an witness.
	//
	// A witness can vote in elections but it doesn't have any Raft log or
	// application state machine associated. The witness node can not be used
	// to initiate read, write or membership change operations on its Raft cluster.
	// Section 11.7.2 of Diego Ongaro's thesis contains more info on such witness
	// role.
	//
	// Application should later call StartCluster with config.Config.IsWitness
	// set to true on the right NodeHost to actually start the witness node.
	//
	// By default, the target parameter is the RaftAddress of the NodeHost instance
	// where the new Raft node will be running. Note that fixed IP or static DNS
	// name should be used in RaftAddress in such default mode. When running in the
	// AddressByNodeHostID mode, target should be set to NodeHost's ID value which
	// can be obtained by calling the ID() method.
	//
	// When the Raft cluster is created with the OrderedConfigChange config flag
	// set as false, the configChangeIndex parameter is ignored. Otherwise, it
	// should be set to the most recent Config Change Index value returned by the
	// SyncGetClusterMembership method. The requested add witness operation will be
	// rejected if other membership change has been applied since that earlier call
	// to the SyncGetClusterMembership method.
	RequestAddWitness(clusterID uint64, nodeID uint64, target dgbclt.Target, configChangeIndex uint64, timeout time.Duration) (*dgbclt.RequestState, error)

	// RequestLeaderTransfer makes a request to transfer the leadership of the
	// specified Raft cluster to the target node identified by targetNodeID. It
	// returns an error if the request fails to be started. There is no guarantee
	// that such request can be fulfilled, i.e. the leadership transfer can still
	// fail after a successful return of the RequestLeaderTransfer method.
	RequestLeaderTransfer(clusterID uint64, targetNodeID uint64) error

	// SyncRemoveData is the synchronous variant of the RemoveData. It waits for
	// the specified node to be fully offloaded or until the ctx instance is
	// cancelled or timeout.
	//
	// Similar to RemoveData, calling SyncRemoveData on a node that is still a Raft
	// cluster member will corrupt the Raft cluster.
	SyncRemoveData(ctx context.Context, clusterID uint64, nodeID uint64) error

	// RemoveData tries to remove all data associated with the specified node. This
	// method should only be used after the node has been deleted from its Raft
	// cluster. Calling RemoveData on a node that is still a Raft cluster member
	// will corrupt the Raft cluster.
	//
	// RemoveData returns ErrClusterNotStopped when the specified node has not been
	// fully offloaded from the NodeHost instance.
	RemoveData(clusterID uint64, nodeID uint64) error

	// GetNodeUser returns an INodeUser instance ready to be used to directly make
	// proposals or read index operations without locating the node repeatedly in
	// the NodeHost. A possible use case is when loading a large data set say with
	// billions of proposals into the dragonboat based system.
	GetNodeUser(clusterID uint64) (dgbclt.INodeUser, error)

	// HasNodeInfo returns a boolean value indicating whether the specified node
	// has been bootstrapped on the current NodeHost instance.
	HasNodeInfo(clusterID uint64, nodeID uint64) bool

	// GetNodeHostInfo returns a NodeHostInfo instance that contains all details
	// of the NodeHost, this includes details of all Raft clusters managed by the
	// the NodeHost instance.
	GetNodeHostInfo(opt dgbclt.NodeHostInfoOption) *dgbclt.NodeHostInfo
}

func NewCluster(cfg *dgbcfg.NodeHostConfig) (Cluster, error) {
	if c, e := dgbclt.NewNodeHost(*cfg); e != nil {
		return nil, e
	} else {
		return &cRaft{
			c: c,
		}, nil
	}
}
