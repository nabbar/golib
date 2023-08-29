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

package nats

import (
	"net/url"
	"os"
	"time"

	libtls "github.com/nabbar/golib/certificates"
	natjwt "github.com/nats-io/jwt/v2"
	natsrv "github.com/nats-io/nats-server/v2/server"
)

type ConfigCustom struct {
	AccountResolver            natsrv.AccountResolver `mapstructure:"-" json:"-" yaml:"-" toml:"-"`
	AccountResolverTLS         bool                   `mapstructure:"-" json:"-" yaml:"-" toml:"-"`
	AccountResolverTLSConfig   libtls.Config          `mapstructure:"-" json:"-" yaml:"-" toml:"-"`
	CustomClientAuthentication natsrv.Authentication  `mapstructure:"-" json:"-" yaml:"-" toml:"-"`
	CustomRouterAuthentication natsrv.Authentication  `mapstructure:"-" json:"-" yaml:"-" toml:"-"`
}

// ConfigNkey is for multiple nkey based users.
type ConfigNkey struct {
	//Nkey is a new challenge introduced by NATS v2 (ED25519 keys).
	Nkey string `mapstructure:"user" json:"user" yaml:"user" toml:"user"`

	//Account define the account associated to this NKey.
	Account string `mapstructure:"account" json:"account" yaml:"account" toml:"account"`

	// SigningKey define the ED 25519 signingKey.
	SigningKey string `mapstructure:"signing_key" json:"signing_key" yaml:"signing_key" toml:"signing_key"`

	//AllowedConnectionTypes define a list of allowed connection, in list of : STANDARD, WEBSOCKET, LEAFNODE, MQTT.
	AllowedConnectionTypes []string `mapstructure:"connection_types" json:"connection_types" yaml:"connection_types" toml:"connection_types"`
}

// ConfigUser is for multiple accounts/users.
type ConfigUser struct {
	//Username is the username used for connection.
	Username string `mapstructure:"username" json:"username" yaml:"username" toml:"username"`

	//Password define the password used for connection.
	Password string `mapstructure:"password" json:"password" yaml:"password" toml:"password"`

	//Account define the account associated to this NKey.
	Account string `mapstructure:"account" json:"account" yaml:"account" toml:"account"`

	//AllowedConnectionTypes define a list of allowed connection, in list of : STANDARD, WEBSOCKET, LEAFNODE, MQTT.
	AllowedConnectionTypes []string `mapstructure:"connection_types" json:"connection_types" yaml:"connection_types" toml:"connection_types"`
}

// ConfigPermissionsUser are the allowed subjects on a per publish or subscribe basis.
type ConfigPermissionsUser struct {
	//Publish define the scope permission for publisher role.
	Publish ConfigPermissionSubject `mapstructure:"publish" json:"publish" yaml:"publish" toml:"publish" validate:"required"`

	//Subscribe define the scope permission for subscriber role.
	Subscribe ConfigPermissionSubject `mapstructure:"subscribe" json:"subscribe" yaml:"subscribe" toml:"subscribe" validate:"required"`

	//Response define the scope permission to allow response for a message.
	Response ConfigPermissionResponse `mapstructure:"response" json:"response" yaml:"response" toml:"response" validate:"required"`
}

// ConfigPermissionsRoute are similar to user permissions but describe what a server can import/export from and to another server.
type ConfigPermissionsRoute struct {
	//Import define the scope permission to import data from the route.
	Import ConfigPermissionSubject `mapstructure:"import" json:"import" yaml:"import" toml:"import" validate:"required"`

	//Export define the scope permission to export data to the route.
	Export ConfigPermissionSubject `mapstructure:"export" json:"export" yaml:"export" toml:"export" validate:"required"`
}

// ConfigPermissionResponse can be used to allow responses to any reply subject that is received on a valid subscription.
type ConfigPermissionResponse struct {
	//MaxMsgs define the maximum message response in the expire duration.
	MaxMsgs int `mapstructure:"max_msgs" json:"max_msgs" yaml:"max_msgs" toml:"max_msgs"`

	//Expires define the TTL of the limitation for max messages.
	Expires time.Duration `mapstructure:"expires" json:"expires" yaml:"expires" toml:"expires"`
}

// ConfigPermissionSubject is an individual allow and deny struct for publish and subscribe authorizations.
type ConfigPermissionSubject struct {
	//Allow define the allowed scope for permission.
	Allow []string `mapstructure:"allow" json:"allow" yaml:"allow" toml:"allow"`

	//Deny define the deny scope for permission.
	Deny []string `mapstructure:"deny" json:"deny" yaml:"deny" toml:"deny"`
}

// ConfigAccount are subject namespace definitions. By default no messages are shared between accounts.
// You can share via Exports and Imports of Streams and Services.
type ConfigAccount struct {
	//Name define the name of the account.
	Name string `mapstructure:"name" json:"name" yaml:"name" toml:"name"`

	Permission ConfigPermissionsUser `mapstructure:"permission" json:"permission" yaml:"permission" toml:"permission" validate:"required"`
}

type ConfigAuth struct {
	//NKeys Set the nkeys list with account.
	NKeys []ConfigNkey `mapstructure:"nkeys" json:"nkeys" yaml:"nkeys" toml:"nkeys" validate:"dive"`

	//Users Set the users list with account.
	Users []ConfigUser `mapstructure:"users" json:"users" yaml:"users" toml:"users" validate:"dive"`

	//Account Set the account list with permissions.
	Accounts []ConfigAccount `mapstructure:"accounts" json:"accounts" yaml:"accounts" toml:"accounts" validate:"dive"`

	//AuthTimeout define the timeout for authentication process.
	AuthTimeout time.Duration `mapstructure:"auth_timeout" json:"auth_timeout" yaml:"auth_timeout" toml:"auth_timeout"`

	//NoAuthUser allows you to refer to a configured user/account when no credentials are provided.
	NoAuthUser string `mapstructure:"no_auth_user" json:"no_auth_user" yaml:"no_auth_user" toml:"no_auth_user"`

	//SystemAccount define the account under which nats-server offer services.
	SystemAccount string `mapstructure:"system_account" json:"system_account" yaml:"system_account" toml:"system_account"`

	//NoSystemAccount disable the system account.
	NoSystemAccount bool `mapstructure:"no_system_account" json:"no_system_account" yaml:"no_system_account" toml:"no_system_account"`

	//TrustedKeys define the list of trusted keys allowed to operate the server
	TrustedKeys []string `mapstructure:"trusted_keys" json:"trusted_keys" yaml:"trusted_keys" toml:"trusted_keys"`

	//TrustedOperators define a list of jwt file for operator claim.
	TrustedOperators []string `mapstructure:"trusted_operators" json:"trusted_operators" yaml:"trusted_operators" toml:"trusted_operators"`
}

type ConfigLogger struct {
	//LogFile define the file to store log output.
	LogFile string `mapstructure:"log_file" json:"log_file" yaml:"log_file" toml:"log_file"`

	// PermissionFolderLogFile is the permission apply if a folder is created
	PermissionFolderLogFile os.FileMode `mapstructure:"permission_folder" json:"permission_folder" yaml:"permission_folder" toml:"permission_folder"`

	// PermissionFileLogFile is the permission apply if a file is created
	PermissionFileLogFile os.FileMode `mapstructure:"permission_file" json:"permission_file" yaml:"permission_file" toml:"permission_file"`

	//Syslog define if log output must be sent to syslog.
	Syslog bool `mapstructure:"syslog" json:"syslog" yaml:"syslog" toml:"syslog"`

	//RemoteSyslog define the syslog server address like '(udp://127.0.0.1:514)'.
	RemoteSyslog string `mapstructure:"remote_syslog" json:"remote_syslog" yaml:"remote_syslog" toml:"remote_syslog"`

	//LogSizeLimit define the maximum size allowed for the log file.
	LogSizeLimit int64 `mapstructure:"log_size_limit" json:"log_size_limit" yaml:"log_size_limit" toml:"log_size_limit"`

	//MaxTracedMsgLen define the max size in chars of trace message
	MaxTracedMsgLen int `mapstructure:"max_traced_msg_len" json:"max_traced_msg_len" yaml:"max_traced_msg_len" toml:"max_traced_msg_len"`

	// ConnectErrorReports specifies the number of failed attempts at which point server should report the failure of an initial connection to a route, gateway or leaf node.
	// See DEFAULT_CONNECT_ERROR_REPORTS for default value.
	ConnectErrorReports int `mapstructure:"connect_error_reports" json:"connect_error_reports" yaml:"connect_error_reports" toml:"connect_error_reports"`

	// ReconnectErrorReports is similar to ConnectErrorReports except that this applies to reconnect events.
	ReconnectErrorReports int `mapstructure:"reconnect_error_reports" json:"reconnect_error_reports" yaml:"reconnect_error_reports" toml:"reconnect_error_reports"`
}

type ConfigLimits struct {
	//MaxConn Set maximum connection.
	MaxConn int `mapstructure:"max_conn" json:"max_conn" yaml:"max_conn" toml:"max_conn"`

	//MaxSubs Set the maximum subscriptions.
	MaxSubs int `mapstructure:"max_subs" json:"max_subs" yaml:"max_subs" toml:"max_subs"`

	//PingInterval define a duration between 2 ping with cluster.
	PingInterval time.Duration `mapstructure:"ping_interval" json:"ping_interval" yaml:"ping_interval" toml:"ping_interval"`

	//MaxPingsOut define the number of ping error before closing connection.
	MaxPingsOut int `mapstructure:"port" json:"max_pings_out" yaml:"port" toml:"port"`

	//MaxControlLine define the maximum allowed protocol control line size.
	MaxControlLine int `mapstructure:"max_control_line" json:"max_control_line" yaml:"max_control_line" toml:"portmax_control_line"`

	//MaxPayload define the maximum allowed payload size.
	MaxPayload int `mapstructure:"max_payload" json:"max_payload" yaml:"max_payload" toml:"max_payload"`

	//MaxPending define the maximum outbound pending bytes per client.
	MaxPending int64 `mapstructure:"max_pending" json:"max_pending" yaml:"max_pending" toml:"max_pending"`

	//WriteDeadline define the deadline timeout to flush line of stream message.
	WriteDeadline time.Duration `mapstructure:"write_deadline" json:"write_deadline" yaml:"write_deadline" toml:"write_deadline"`

	//MaxClosedClients define the number of closed clients connection to keep.
	MaxClosedClients int `mapstructure:"max_closed_clients" json:"max_closed_clients" yaml:"max_closed_clients" toml:"max_closed_clients"`

	//LameDuckDuration define the timeout to closing all client in LameDuck mode with signal ldm.
	LameDuckDuration time.Duration `mapstructure:"lame_duck_duration" json:"lame_duck_duration" yaml:"lame_duck_duration" toml:"lame_duck_duration"`

	//LameDuckGracePeriod define the grace period before closing client connection for LameDuck shutdown mode.
	LameDuckGracePeriod time.Duration `mapstructure:"lame_duck_grace_period" json:"lame_duck_grace_period" yaml:"lame_duck_grace_period" toml:"lame_duck_grace_period"`

	//NoSublistCache define the option to disable subscription caches for all accounts.
	//This is saves resources in situations where different subjects are used all the time.
	NoSublistCache bool `mapstructure:"no_sublist_cache" json:"no_sublist_cache" yaml:"no_sublist_cache" toml:"no_sublist_cache"`

	//NoHeaderSupport define the option to disable header in the server.
	//No use except for nats.js, nats.ws, nats.deno and docker image nighty build.
	NoHeaderSupport bool `mapstructure:"no_header_support" json:"no_header_support" yaml:"no_header_support" toml:"no_header_support"`

	//DisableShortFirstPing define the option to disable the very first PING to a lower interval to capture the initial RTT.
	// After that the PING interval will be set to the user defined value.
	DisableShortFirstPing bool `mapstructure:"disable_short_first_ping" json:"disable_short_first_ping" yaml:"disable_short_first_ping" toml:"disable_short_first_ping"`
}

type ConfigSrv struct {
	//Name define the name of the server.
	Name string `mapstructure:"name" json:"name" yaml:"name" toml:"name"`

	//Host define the network host to listen on.
	Host string `mapstructure:"host" json:"host" yaml:"host" toml:"host"`

	//Port define the network port to listen on.
	Port int `mapstructure:"port" json:"port" yaml:"port" toml:"port"`

	//ClientAdvertise is an alternative client listen specification <host>:<port> or just <host> to
	//advertise to clients and other server. Useful in cluster setups with NAT.
	ClientAdvertise string `mapstructure:"client_advertise" json:"client_advertise" yaml:"client_advertise" toml:"client_advertise"`

	//HTTPHost define host use to expose monitoring api.
	HTTPHost string `mapstructure:"http_host" json:"http_host" yaml:"http_host" toml:"http_host"`

	//HTTPPort define port use to expose monitoring api.
	HTTPPort int `mapstructure:"http_port" json:"http_port" yaml:"http_port" toml:"http_port"`

	//HTTPPort define port use to expose monitoring api with tls.
	HTTPSPort int `mapstructure:"https_port" json:"https_port" yaml:"https_port" toml:"https_port"`

	//HTTPBasePath define the base path for monitoring endpoints.
	HTTPBasePath string `mapstructure:"http_base_path" json:"http_base_path" yaml:"http_base_path" toml:"http_base_path"`

	//ProfPort define the Profiling HTTP port to enable server for dynamic profiling.
	ProfPort int `mapstructure:"prof_port" json:"prof_port" yaml:"prof_port" toml:"prof_port"`

	//PidFile define the file path to store PID process.
	PidFile string `mapstructure:"pid_file" json:"pid_file" yaml:"pid_file" toml:"pid_file"`

	//PortsFileDir define the directory where ports file will be created like '<executable_name>_<pid>.ports'.
	PortsFileDir string `mapstructure:"ports_file_dir" json:"ports_file_dir" yaml:"ports_file_dir" toml:"ports_file_dir"`

	//Routes define a list of url to actively solicit a connection.
	Routes []*url.URL `mapstructure:"routes" json:"routes" yaml:"routes" toml:"routes"`

	//RoutesStr define the routes to actively solicit a connection.
	RoutesStr string `mapstructure:"routes_str" json:"routes_str" yaml:"routes_str" toml:"routes_str"`

	//NoSig is used to disable signal catch for server.
	NoSig bool `mapstructure:"no_log" json:"no_log" yaml:"no_log" toml:"no_log"`

	//Username is the username used for server connection (like flag --auth in NATS v2 documentation).
	Username string `mapstructure:"username" json:"username" yaml:"username" toml:"username"`

	//Password define the password used for server connection (like flag --auth in NATS v2 documentation).
	Password string `mapstructure:"password" json:"password" yaml:"password" toml:"password"`

	//Token define the token used for server connection (like flag --auth in NATS v2 documentation).
	Token string `mapstructure:"token" json:"token" yaml:"token" toml:"token"`

	//JetStream allow to enable or disable jetStream layer
	JetStream bool `mapstructure:"jet_stream" json:"jet_stream" yaml:"jet_stream" toml:"jet_stream"`

	//JetStreamMaxMemory define the maximum memory used for jetStream in memory store type
	JetStreamMaxMemory int64 `mapstructure:"jet_stream_max_memory" json:"jet_stream_max_memory" yaml:"jet_stream_max_memory" toml:"jet_stream_max_memory"`

	//JetStreamMaxStore define the maximum disk used for jetStream in file store type
	JetStreamMaxStore int64 `mapstructure:"jet_stream_max_store" json:"jet_stream_max_store" yaml:"jet_stream_max_store" toml:"jet_stream_max_store"`

	//StoreDir define the directory path for jetStream in file store type
	StoreDir string `mapstructure:"store_dir" json:"store_dir" yaml:"store_dir" toml:"store_dir"`

	// PermissionStoreDir is the permission apply if a folder is created
	PermissionStoreDir os.FileMode `mapstructure:"permission_store_dir" json:"permission_store_dir" yaml:"permission_store_dir" toml:"permission_store_dir"`

	//Tags describing the server.
	//They will be included in varz and used as a filter criteria for some system requests
	Tags natjwt.TagList `mapstructure:"tags" json:"tags" yaml:"tags" toml:"tags" validate:"dive"`

	//TLS Enable tls for server.
	TLS bool `mapstructure:"tls" json:"tls" yaml:"tls" toml:"tls"`

	//AllowNoTLS define if client no TLS connection are allowed or no.
	AllowNoTLS bool `mapstructure:"allow_no_tls" json:"allow_no_tls" yaml:"allow_no_tls" toml:"allow_no_tls"`

	//TLSTimeout define the timeout for tls handshake for client and http monitoring.
	TLSTimeout time.Duration `mapstructure:"tls_timeout" json:"tls_timeout" yaml:"tls_timeout" toml:"tls_timeout"`

	//TLSConfig Configuration map for tls for client and http monitoring.
	TLSConfig libtls.Config `mapstructure:"tls_config" json:"tls_config" yaml:"tls_config" toml:"tls_config" validate:""`
}

type ConfigCluster struct {
	//Name define the name of the cluster.
	Name string `mapstructure:"name" json:"name" yaml:"name" toml:"name"`

	//Host define the network host to cluster listen on.
	Host string `mapstructure:"host" json:"host" yaml:"host" toml:"host"`

	//Port define the network port to cluster listen on.
	Port int `mapstructure:"port" json:"port" yaml:"port" toml:"port"`

	//ListenStr define the cluster url from which members can solicit routes.
	ListenStr string `mapstructure:"listen_str" json:"listen_str" yaml:"listen_str" toml:"listen_str"`

	//Advertise define the cluster URL to advertise to other servers.
	Advertise string `mapstructure:"advertise" json:"advertise" yaml:"advertise" toml:"advertise"`

	//NoAdvertise specify if Advertise known cluster IPs to clients.
	NoAdvertise bool `mapstructure:"no_advertise" json:"no_advertise" yaml:"no_advertise" toml:"no_advertise"`

	//ConnectRetries define the number of connect retries for implicit routes.
	ConnectRetries int `mapstructure:"connect_retries" json:"connect_retries" yaml:"connect_retries" toml:"connect_retries"`

	//Username is the username used for cluster connection.
	Username string `mapstructure:"username" json:"username" yaml:"username" toml:"username"`

	//Password define the password used for cluster connection.
	Password string `mapstructure:"password" json:"password" yaml:"password" toml:"password"`

	//AuthTimeout define the timeout for authentication process.
	AuthTimeout time.Duration `mapstructure:"auth_timeout" json:"auth_timeout" yaml:"auth_timeout" toml:"auth_timeout"`

	//Permissions define the scope permission assign to route connections.
	Permissions ConfigPermissionsRoute `mapstructure:"permissions" json:"permissions" yaml:"permissions" toml:"permissions" validate:""`

	//TLS Enable tls for cluster connection.
	TLS bool `mapstructure:"tls" json:"tls" yaml:"tls" toml:"tls"`

	//TLSTimeout define the timeout for tls handshake for cluster connection.
	TLSTimeout time.Duration `mapstructure:"tls_timeout" json:"tls_timeout" yaml:"tls_timeout" toml:"tls_timeout"`

	//TLSConfig define the tls configuration for cluster connection.
	TLSConfig libtls.Config `mapstructure:"tls_config" json:"tls_config" yaml:"tls_config" toml:"tls_config" validate:""`
}

// ConfigGatewayRemote are options for connecting to a remote gateway
// NOTE: This structure is no longer used for monitoring endpoints and json tags are deprecated and may be removed in the future.
type ConfigGatewayRemote struct {
	//Name define the name of the current gateways destination.
	Name string `mapstructure:"name" json:"name" yaml:"name" toml:"name"`

	//URLs define a list of route for the current gateways destination.
	URLs []*url.URL `mapstructure:"urls" json:"urls" yaml:"urls" toml:"urls" validate:"dive"`

	//TLS Enable tls for the current gateways destination.
	TLS bool `mapstructure:"tls" json:"tls" yaml:"tls" toml:"tls"`

	//TLSTimeout define the timeout for tls handshake for the current gateways destination.
	TLSTimeout time.Duration `mapstructure:"tls_timeout" json:"tls_timeout" yaml:"tls_timeout" toml:"tls_timeout"`

	//TLSConfig define the tls configuration for the current gateways destination.
	TLSConfig libtls.Config `mapstructure:"tls_config" json:"tls_config" yaml:"tls_config" toml:"tls_config" validate:""`
}

// ConfigGateway are options for gateways.
// NOTE: This structure is no longer used for monitoring endpoints and json tags are deprecated and may be removed in the future.
type ConfigGateway struct {
	//Name define the name of the gateway.
	Name string `mapstructure:"name" json:"name" yaml:"name" toml:"name"`

	//Host define the network host to listen on.
	Host string `mapstructure:"host" json:"host" yaml:"host" toml:"host"`

	//Port define the network port to listen on.
	Port int `mapstructure:"port" json:"port" yaml:"port" toml:"port"`

	//Username is the username used for gateways connection.
	Username string `mapstructure:"username" json:"username" yaml:"username" toml:"username"`

	//Password define the password used for gateways connection.
	Password string `mapstructure:"password" json:"password" yaml:"password" toml:"password"`

	//AuthTimeout define the timeout for authentication process.
	AuthTimeout time.Duration `mapstructure:"auth_timeout" json:"auth_timeout" yaml:"auth_timeout" toml:"auth_timeout"`

	//Advertise define the gateway URL to advertise to other servers.
	Advertise string `mapstructure:"advertise" json:"advertise" yaml:"advertise" toml:"advertise"`

	//ConnectRetries define the number of connect retries for implicit routes.
	ConnectRetries int `mapstructure:"connect_retries" json:"connect_retries" yaml:"connect_retries" toml:"connect_retries"`

	//Gateways define a list of route for gateways.
	Gateways []*ConfigGatewayRemote `mapstructure:"gateways" json:"gateways" yaml:"gateways" toml:"gateways" validate:"dive"`

	//RejectUnknown allow to reject unknown cluster connection.
	RejectUnknown bool `mapstructure:"reject_unknown" json:"reject_unknown" yaml:"reject_unknown" toml:"reject_unknown"`

	//TLS Enable tls for gateways connection.
	TLS bool `mapstructure:"tls" json:"tls" yaml:"tls" toml:"tls"`

	//TLSTimeout define the timeout for tls handshake for gateways connection.
	TLSTimeout time.Duration `mapstructure:"tls_timeout" json:"tls_timeout" yaml:"tls_timeout" toml:"tls_timeout"`

	//TLSConfig define the tls configuration for gateways connection.
	TLSConfig libtls.Config `mapstructure:"tls_config" json:"tls_config" yaml:"tls_config" toml:"tls_config" validate:""`
}

// ConfigLeaf are options for a given server to accept leaf node connections and/or connect to a remote cluster.
type ConfigLeaf struct {
	//Host define the network host to listen on.
	Host string `mapstructure:"host" json:"host" yaml:"host" toml:"host"`

	//Port define the network port to listen on.
	Port int `mapstructure:"port" json:"port" yaml:"port" toml:"port"`

	//Username is the username used for leaf connection.
	Username string `mapstructure:"username" json:"username" yaml:"username" toml:"username"`

	//Password define the password used for leaf connection.
	Password string `mapstructure:"password" json:"password" yaml:"password" toml:"password"`

	//AuthTimeout define the timeout for authentication process.
	AuthTimeout time.Duration `mapstructure:"auth_timeout" json:"auth_timeout" yaml:"auth_timeout" toml:"auth_timeout"`

	//Advertise define the gateway URL to advertise to other servers.
	Advertise string `mapstructure:"advertise" json:"advertise" yaml:"advertise" toml:"advertise"`

	//NoAdvertise specify if Advertise known leaf node IPs to clients.
	NoAdvertise bool `mapstructure:"no_advertise" json:"no_advertise" yaml:"no_advertise" toml:"no_advertise"`

	//Account define the account under which leaf offer services.
	Account string `mapstructure:"account" json:"account" yaml:"account" toml:"account"`

	//Users Set the users list with account.
	Users []ConfigUser `mapstructure:"users" json:"users" yaml:"users" toml:"users" validate:"dive"`

	//ReconnectInterval define the duration to wait after a closed connection and before trying to reconnect
	ReconnectInterval time.Duration `mapstructure:"reconnect_interval" json:"reconnect_interval" yaml:"reconnect_interval" toml:"reconnect_interval"`

	//Remotes define For solicited connections to other clusters/superclusters.
	Remotes []*ConfigLeafRemote `mapstructure:"remotes" json:"remotes" yaml:"remotes" toml:"remotes"`

	//TLS Enable tls for leaf connection.
	TLS bool `mapstructure:"tls" json:"tls" yaml:"tls" toml:"tls"`

	//TLSTimeout define the timeout for tls handshake for leaf connection.
	TLSTimeout time.Duration `mapstructure:"tls_timeout" json:"tls_timeout" yaml:"tls_timeout" toml:"tls_timeout"`

	//TLSConfig define the tls configuration for leaf connection.
	TLSConfig libtls.Config `mapstructure:"tls_config" json:"tls_config" yaml:"tls_config" toml:"tls_config"`
}

// ConfigLeafRemote are options for connecting to a remote server as a leaf node.
type ConfigLeafRemote struct {
	//LocalAccount define the local account to use for this remote cluster.
	LocalAccount string `mapstructure:"local_account" json:"local_account" yaml:"local_account" toml:"local_account"`

	//URLs define a list of route for the current gateways destination.
	URLs []*url.URL `mapstructure:"urls" json:"urls" yaml:"urls" toml:"urls"`

	//Credentials define the file path for authentication with credentials file
	Credentials string `mapstructure:"credentials" json:"credentials" yaml:"credentials" toml:"credentials"`

	//Hub define the remote connection as hub
	Hub bool `mapstructure:"hub" json:"hub" yaml:"hub" toml:"hub"`

	//DenyImports define a list of subject/queues denied for import
	DenyImports []string `mapstructure:"deny_imports" json:"deny_imports" yaml:"deny_imports" toml:"deny_imports"`

	//DenyExports define a list of subject/queues denied for export
	DenyExports []string `mapstructure:"deny_exports" json:"deny_exports" yaml:"deny_exports" toml:"deny_exports"`

	//Websocket define options for websocket connections.
	// When an URL has the "ws" (or "wss") scheme, then the server will initiate the
	// connection as a websocket connection. By default, the websocket frames will be
	// masked (as if this server was a websocket client to the remote server). The
	// NoMasking option will change this behavior and will send umasked frames.
	Websocket struct {
		//Compression define if compression is enable for remote ws connection
		Compression bool `mapstructure:"compression" json:"compression" yaml:"compression" toml:"compression"`

		//NoMasking define if the remote ws is must be masked.
		//By default ws are masked but this option allow to expose it.
		NoMasking bool `mapstructure:"no_masking" json:"no_masking" yaml:"no_masking" toml:"no_masking"`
	} `mapstructure:"websocket" json:"websocket" yaml:"websocket" toml:"websocket"`

	//TLS Enable tls for this remote cluster connection.
	TLS bool `mapstructure:"tls" json:"tls" yaml:"tls" toml:"tls"`

	//TLSTimeout define the timeout for tls handshake for this remote cluster connection.
	TLSTimeout time.Duration `mapstructure:"tls_timeout" json:"tls_timeout" yaml:"tls_timeout" toml:"tls_timeout"`

	//TLSConfig define the tls configuration for this remote cluster connection.
	TLSConfig libtls.Config `mapstructure:"tls_config" json:"tls_config" yaml:"tls_config" toml:"tls_config"`
}

// ConfigWebsocket are options for websocket
type ConfigWebsocket struct {

	//Host The server will accept websocket client connections on this hostname/IP.
	Host string `mapstructure:"" json:"host" yaml:"host" toml:"host"`

	//Port The server will accept websocket client connections on this port.
	Port int `mapstructure:"port" json:"port" yaml:"port" toml:"port"`

	//Advertise The host:port to advertise to websocket clients in the cluster.
	Advertise string `mapstructure:"advertise" json:"advertise" yaml:"advertise" toml:"advertise"`

	//NoAuthUser define the default user for new client connection.
	//If no user name is provided when a client connects, will default to the matching user from the global list of users in `Options.Users`.
	NoAuthUser string `mapstructure:"no_auth_user" json:"no_auth_user" yaml:"no_auth_user" toml:"no_auth_user"`

	//JWTCookie define the name of the cookie, which if present in WebSocket upgrade headers,
	//will be treated as JWT during CONNECT phase as long as "jwt" specified in the CONNECT options is missing or empty.
	JWTCookie string `mapstructure:"jwt_cookie" json:"jwt_cookie" yaml:"jwt_cookie" toml:"jwt_cookie"`

	//Authentication section.
	//If anything is configured in this section, it will override the authorization configuration of regular clients.

	//Username is the username used.
	Username string `mapstructure:"username" json:"username" yaml:"username" toml:"username"`

	//Password define the password used.
	Password string `mapstructure:"password" json:"password" yaml:"password" toml:"password"`

	//Token define the token used.
	Token string `mapstructure:"token" json:"token" yaml:"token" toml:"token"`

	//AuthTimeout define the timeout for authentication process.
	AuthTimeout time.Duration `mapstructure:"auth_timeout" json:"auth_timeout" yaml:"auth_timeout" toml:"auth_timeout"`

	//SameOrigin define if request's host must match the Origin header or not.
	SameOrigin bool `mapstructure:"same_origin" json:"same_origin" yaml:"same_origin" toml:"same_origin"`

	//AllowedOrigins define a list of allowed origin could not matching with request's host.
	// Only origins in this list will be accepted. If empty and SameOrigin is false, any origin is accepted.
	AllowedOrigins []string `mapstructure:"allowed_origins" json:"allowed_origins" yaml:"allowed_origins" toml:"allowed_origins"`

	//Compression allow to activate compression between client and server.
	//If set to true, the server will negotiate with clients if compression can be used.
	//If this is false, no compression will be used (both in server and clients) since it has to be negotiated between both endpoints
	Compression bool `mapstructure:"compression" json:"compression" yaml:"compression" toml:"compression"`

	// NoTLS allow to start websocket server without tls.
	NoTLS bool `mapstructure:"no_tls" json:"no_tls" yaml:"no_tls" toml:"no_tls"`

	//HandshakeTimeout is the total time allowed for the server to read the client request and write the response back to the client.
	//This include the time needed for the TLS Handshake.
	HandshakeTimeout time.Duration `mapstructure:"handshake_timeout" json:"handshake_timeout" yaml:"handshake_timeout" toml:"handshake_timeout"`

	//TLSConfig define the tls configuration for this remote cluster connection.
	TLSConfig libtls.Config `mapstructure:"tls_config" json:"tls_config" yaml:"tls_config" toml:"tls_config"`
}

// ConfigMQTT are options for MQTT
type ConfigMQTT struct {

	//Host define the hostname/IP to accept MQTT client connections.
	Host string `mapstructure:"" json:"host" yaml:"host" toml:"host"`

	//Port define the port to accept MQTT client connections.
	Port int `mapstructure:"port" json:"port" yaml:"port" toml:"port"`

	//NoAuthUser define the default user for new client connection.
	//If no user name is provided when a client connects, will default to the matching user from the global list of users in `Options.Users`.
	NoAuthUser string `mapstructure:"no_auth_user" json:"no_auth_user" yaml:"no_auth_user" toml:"no_auth_user"`

	//Authentication section.
	//If anything is configured in this section, it will override the authorization configuration of regular clients.

	//Username is the username used.
	Username string `mapstructure:"username" json:"username" yaml:"username" toml:"username"`

	//Password define the password used.
	Password string `mapstructure:"password" json:"password" yaml:"password" toml:"password"`

	//Token define the token used.
	Token string `mapstructure:"token" json:"token" yaml:"token" toml:"token"`

	//AuthTimeout define the timeout for authentication process.
	AuthTimeout time.Duration `mapstructure:"auth_timeout" json:"auth_timeout" yaml:"auth_timeout" toml:"auth_timeout"`

	// AckWait is the amount of time after which a QoS 1 message sent to
	// a client is redelivered as a DUPLICATE if the server has not
	// received the PUBACK on the original Packet Identifier.
	// The value has to be positive.
	// Zero will cause the server to use the default value (30 seconds).
	// Note that changes to this option is applied only to new MQTT subscriptions.
	AckWait time.Duration `mapstructure:"ack_wait" json:"ack_wait" yaml:"ack_wait" toml:"ack_wait"`

	// MaxAckPending is the amount of QoS 1 messages the server can send to
	// a subscription without receiving any PUBACK for those messages.
	// The valid range is [0..65535].
	// The total of subscriptions' MaxAckPending on a given session cannot
	// exceed 65535. Attempting to create a subscription that would bring
	// the total above the limit would result in the server returning 0x80
	// in the SUBACK for this subscription.
	// Due to how the NATS Server handles the MQTT "#" wildcard, each
	// subscription ending with "#" will use 2 times the MaxAckPending value.
	// Note that changes to this option is applied only to new subscriptions.
	MaxAckPending uint16 `mapstructure:"max_ack_pending" json:"max_ack_pending" yaml:"max_ack_pending" toml:"max_ack_pending"`

	//TLS Enable tls for this remote cluster connection.
	TLS bool `mapstructure:"tls" json:"tls" yaml:"tls" toml:"tls"`

	//TLSTimeout is the time needed for the TLS Handshake.
	TLSTimeout time.Duration `mapstructure:"tls_timeout" json:"tls_timeout" yaml:"tls_timeout" toml:"tls_timeout"`

	//TLSConfig define the tls configuration for MQTT client connection.
	TLSConfig libtls.Config `mapstructure:"tls_config" json:"tls_config" yaml:"tls_config" toml:"tls_config"`
}
