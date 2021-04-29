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
	"time"

	libtls "github.com/nabbar/golib/certificates"
)

// ConfigUser is for multiple accounts/users.
type ConfigUser struct {
	//Username is the username used for connection
	Username string `mapstructure:"username" json:"username" yaml:"username" toml:"username"`

	//Password define the password used for connection
	Password string `mapstructure:"password" json:"password" yaml:"password" toml:"password"`

	//Permissions define the scope permission assign to this user for role publisher and / or subscriber
	Permissions ConfigUserPermissions `mapstructure:"publish" json:"publish" yaml:"publish" toml:"publish"`
}

// ConfigUserPermissions are the allowed subjects on a per publish or subscribe basis.
type ConfigUserPermissions struct {
	//Publish define the scope permission for publisher role
	Publish ConfigSubjectPermission `mapstructure:"publish" json:"publish" yaml:"publish" toml:"publish"`

	//Subscribe define the scope permission for subscriber role
	Subscribe ConfigSubjectPermission `mapstructure:"subscribe" json:"subscribe" yaml:"subscribe" toml:"subscribe"`
}

// ConfigRoutePermissions are similar to user permissions but describe what a server can import/export from and to another server.
type ConfigRoutePermissions struct {
	//Import define the scope permission to import data from the route
	Import ConfigSubjectPermission `mapstructure:"import" json:"import" yaml:"import" toml:"import"`

	//Export define the scope permission to export data to the route
	Export ConfigSubjectPermission `mapstructure:"export" json:"export" yaml:"export" toml:"export"`
}

// ConfigSubjectPermission is an individual allow and deny struct for publish and subscribe authorizations.
type ConfigSubjectPermission struct {
	//Allow define the allowed scope for permission
	Allow []string `mapstructure:"allow" json:"allow" yaml:"allow" toml:"allow"`

	//Deny define the deny scope for permission
	Deny []string `mapstructure:"deny" json:"deny" yaml:"deny" toml:"deny"`
}

type ConfigAuth struct {
	//Users Set the users list with permissions
	Users []ConfigUser `mapstructure:"users" json:"users" yaml:"users" toml:"users"`

	//AuthTimeout define the timeout for authentication process
	AuthTimeout time.Duration `mapstructure:"auth_timeout" json:"auth_timeout" yaml:"auth_timeout" toml:"auth_timeout"`
}

type ConfigLogger struct {
	//NoLog is used to disable log for nats
	NoLog bool `mapstructure:"no_log" json:"no_log" yaml:"no_log" toml:"no_log"`

	//LogFile define the file to store log output
	LogFile string `mapstructure:"log_file" json:"log_file" yaml:"log_file" toml:"log_file"`

	//Syslog define if log output must be sent to syslog
	Syslog bool `mapstructure:"syslog" json:"syslog" yaml:"syslog" toml:"syslog"`

	//RemoteSyslog define the syslog server address like '(udp://127.0.0.1:514)'
	RemoteSyslog string `mapstructure:"remote_syslog" json:"remote_syslog" yaml:"remote_syslog" toml:"remote_syslog"`
}

type ConfigLimits struct {
	//MaxConn Set maximum connection
	MaxConn int `mapstructure:"max_conn" json:"max_conn" yaml:"max_conn" toml:"max_conn"`

	//MaxSubs Set the maximum subscriptions
	MaxSubs int `mapstructure:"max_subs" json:"max_subs" yaml:"max_subs" toml:"max_subs"`

	//PingInterval define a duration between 2 ping with cluster
	PingInterval time.Duration `mapstructure:"ping_interval" json:"ping_interval" yaml:"ping_interval" toml:"ping_interval"`

	//MaxPingsOut define the number of ping error before closing connection
	MaxPingsOut int `mapstructure:"port" json:"max_pings_out" yaml:"port" toml:"port"`

	//MaxControlLine define the maximum allowed protocol control line size
	MaxControlLine int `mapstructure:"max_control_line" json:"max_control_line" yaml:"max_control_line" toml:"portmax_control_line"`

	//MaxPayload define the maximum allowed payload size
	MaxPayload int `mapstructure:"max_payload" json:"max_payload" yaml:"max_payload" toml:"max_payload"`

	//MaxPending define the maximum outbound pending bytes per client
	MaxPending int64 `mapstructure:"max_pending" json:"max_pending" yaml:"max_pending" toml:"max_pending"`

	//WriteDeadline define the deadline timeout to flush line of stream message
	WriteDeadline time.Duration `mapstructure:"write_deadline" json:"write_deadline" yaml:"write_deadline" toml:"write_deadline"`

	//RQSubsSweep define the duration to keep Queues Subs before cleaning
	RQSubsSweep time.Duration `mapstructure:"rq_subs_sweep" json:"rq_subs_sweep" yaml:"rq_subs_sweep" toml:"rq_subs_sweep"`

	//MaxClosedClients define the number of closed clients connection to keep
	MaxClosedClients int `mapstructure:"max_closed_clients" json:"max_closed_clients" yaml:"max_closed_clients" toml:"max_closed_clients"`

	//LameDuckDuration define the timeout to closing all client in LameDuck mode with signal ldm
	LameDuckDuration time.Duration `mapstructure:"lame_duck_duration" json:"lame_duck_duration" yaml:"lame_duck_duration" toml:"lame_duck_duration"`
}

type ConfigSrv struct {
	//Host define the network host to listen on
	Host string `mapstructure:"host" json:"host" yaml:"host" toml:"host"`

	//Port define the network port to listen on
	Port int `mapstructure:"port" json:"port" yaml:"port" toml:"port"`

	//ClientAdvertise is an alternative client listen specification <host>:<port> or just <host> to
	//advertise to clients and other server. Useful in cluster setups with NAT.
	ClientAdvertise string `mapstructure:"client_advertise" json:"client_advertise" yaml:"client_advertise" toml:"client_advertise"`

	//HTTPHost define host use to expose api server
	HTTPHost string `mapstructure:"http_host" json:"http_host" yaml:"http_host" toml:"http_host"`

	//HTTPPort define port use to expose api server
	HTTPPort int `mapstructure:"http_port" json:"http_port" yaml:"http_port" toml:"http_port"`

	//HTTPPort define port use to expose api server with tls
	HTTPSPort int `mapstructure:"https_port" json:"https_port" yaml:"https_port" toml:"https_port"`

	//ProfPort define the Profiling HTTP port to enable server for dynamic profiling
	ProfPort int `mapstructure:"prof_port" json:"prof_port" yaml:"prof_port" toml:"prof_port"`

	//PidFile define the file path to store PID process
	PidFile string `mapstructure:"pid_file" json:"pid_file" yaml:"pid_file" toml:"pid_file"`

	//PortsFileDir define the directory where ports file will be created like '<executable_name>_<pid>.ports'
	PortsFileDir string `mapstructure:"ports_file_dir" json:"ports_file_dir" yaml:"ports_file_dir" toml:"ports_file_dir"`

	//Routes define a list of url to actively solicit a connection
	Routes []*url.URL `mapstructure:"routes" json:"routes" yaml:"routes" toml:"routes"`

	//RoutesStr define the routes to actively solicit a connection
	RoutesStr string `mapstructure:"routes_str" json:"routes_str" yaml:"routes_str" toml:"routes_str"`

	//NoSig is used to disable signal catch for server
	NoSig bool `mapstructure:"no_log" json:"no_log" yaml:"no_log" toml:"no_log"`

	//TLS Enable tls for server
	TLS bool `mapstructure:"tls" json:"tls" yaml:"tls" toml:"tls"`

	//TLSTimeout define the timeout for tls handshake for client and http monitoring
	TLSTimeout time.Duration `mapstructure:"tls_timeout" json:"tls_timeout" yaml:"tls_timeout" toml:"tls_timeout"`

	//TLSConfig Configuration map for tls for client and http monitoring.
	TLSConfig libtls.Config `mapstructure:"tls_config" json:"tls_config" yaml:"tls_config" toml:"tls_config"`
}

type ConfigCluster struct {
	//Host define the network host to cluster listen on
	Host string `mapstructure:"host" json:"host" yaml:"host" toml:"host"`

	//Port define the network port to cluster listen on
	Port int `mapstructure:"port" json:"port" yaml:"port" toml:"port"`

	//ListenStr define the cluster url from which members can solicit routes.
	ListenStr string `mapstructure:"listen_str" json:"listen_str" yaml:"listen_str" toml:"listen_str"`

	//Advertise define the cluster URL to advertise to other servers.
	Advertise string `mapstructure:"advertise" json:"advertise" yaml:"advertise" toml:"advertise"`

	//NoAdvertise specify if Advertise known cluster IPs to clients.
	NoAdvertise bool `mapstructure:"no_advertise" json:"no_advertise" yaml:"no_advertise" toml:"no_advertise"`

	//ConnectRetries define the number of connect retries for implicit routes.
	ConnectRetries int `mapstructure:"connect_retries" json:"connect_retries" yaml:"connect_retries" toml:"connect_retries"`

	//Username is the username used for cluster connection
	Username string `mapstructure:"username" json:"username" yaml:"username" toml:"username"`

	//Password define the password used for cluster connection
	Password string `mapstructure:"password" json:"password" yaml:"password" toml:"password"`

	//AuthTimeout define the timeout for authentication process
	AuthTimeout time.Duration `mapstructure:"auth_timeout" json:"auth_timeout" yaml:"auth_timeout" toml:"auth_timeout"`

	//Permissions define the scope permission assign to route connections
	Permissions ConfigRoutePermissions `mapstructure:"permissions" json:"permissions" yaml:"permissions" toml:"permissions"`

	//TLS Enable tls for cluster connection
	TLS bool `mapstructure:"tls" json:"tls" yaml:"tls" toml:"tls"`

	//TLSTimeout define the timeout for tls handshake for cluster connection
	TLSTimeout time.Duration `mapstructure:"tls_timeout" json:"tls_timeout" yaml:"tls_timeout" toml:"tls_timeout"`

	//TLSConfig define the tls configuration for cluster connection.
	TLSConfig libtls.Config `mapstructure:"tls_config" json:"tls_config" yaml:"tls_config" toml:"tls_config"`
}
