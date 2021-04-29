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
	"fmt"
	"net/url"
	"time"

	"github.com/go-playground/validator/v10"
	libtls "github.com/nabbar/golib/certificates"
	liberr "github.com/nabbar/golib/errors"
	liblog "github.com/nabbar/golib/logger"
	natsrv "github.com/nats-io/nats-server/server"
)

type Config struct {
	Server  ConfigSrv     `mapstructure:"server" json:"server" yaml:"server" toml:"server"`
	Cluster ConfigCluster `mapstructure:"cluster" json:"cluster" yaml:"cluster" toml:"cluster"`
	Limits  ConfigLimits  `mapstructure:"limits" json:"limits" yaml:"limits" toml:"limits"`
	Logs    ConfigLogger  `mapstructure:"logs" json:"logs" yaml:"logs" toml:"logs"`
	Auth    ConfigAuth    `mapstructure:"auth" json:"auth" yaml:"auth" toml:"auth"`
}

func (c Config) Validate() liberr.Error {
	val := validator.New()
	err := val.Struct(c)

	if e, ok := err.(*validator.InvalidValidationError); ok {
		return ErrorValidateConfig.ErrorParent(e)
	}

	out := ErrorValidateConfig.Error(nil)

	for _, e := range err.(validator.ValidationErrors) {
		//nolint goerr113
		out.AddParent(fmt.Errorf("config field '%s' is not validated by constraint '%s'", e.Field(), e.ActualTag()))
	}

	if out.HasParent() {
		return out
	}

	return nil
}

func (c Config) NatsOption(defaultTls libtls.TLSConfig) (*natsrv.Options, liberr.Error) {
	cfg := &natsrv.Options{}

	// Server

	if c.Server.Host != "" {
		cfg.Host = c.Server.Host
	}

	if c.Server.Port > 0 {
		cfg.Port = c.Server.Port
	}

	if c.Server.ClientAdvertise != "" {
		cfg.ClientAdvertise = c.Server.ClientAdvertise
	}

	if c.Server.TLS {
		cfg.TLS = true
		if t, e := c.Server.TLSConfig.NewFrom(defaultTls); e != nil {
			return nil, e
		} else {
			cfg.TLSConfig = t.TlsConfig("")
		}
		if c.Server.TLSTimeout > 0 {
			cfg.TLSTimeout = float64(c.Server.TLSTimeout) / float64(time.Second)
		}
	} else {
		cfg.TLSConfig = nil
		cfg.TLSTimeout = 0
	}

	if c.Server.HTTPHost != "" {
		cfg.HTTPHost = c.Server.HTTPHost
	}

	if c.Server.HTTPPort > 0 {
		cfg.HTTPPort = c.Server.HTTPPort
	}

	if c.Server.HTTPSPort > 0 {
		cfg.HTTPSPort = c.Server.HTTPSPort
	}

	if c.Server.ProfPort > 0 {
		cfg.ProfPort = c.Server.ProfPort
	}

	if c.Server.PidFile != "" {
		cfg.PidFile = c.Server.PidFile
	}

	if c.Server.PortsFileDir != "" {
		cfg.PortsFileDir = c.Server.PortsFileDir
	}

	if len(c.Server.Routes) > 0 {
		cfg.Routes = make([]*url.URL, 0)
		for _, u := range c.Server.Routes {
			if u != nil && u.Host != "" {
				cfg.Routes = append(cfg.Routes, u)
			}
		}
	}

	if c.Server.RoutesStr != "" {
		cfg.RoutesStr = c.Server.RoutesStr
	}

	if c.Server.NoSig {
		cfg.NoSigs = true
	}

	// Logger

	if c.Logs.Syslog {
		cfg.Syslog = true
	}

	if c.Logs.RemoteSyslog != "" {
		cfg.RemoteSyslog = c.Logs.RemoteSyslog
	}

	if c.Logs.LogFile != "" {
		cfg.LogFile = c.Logs.LogFile
	}

	if liblog.IsTimeStamp() {
		cfg.Logtime = true
	}

	if liblog.IsFileTrace() {
		cfg.Trace = true
	}
	switch liblog.GetCurrentLevel() {
	case liblog.DebugLevel:
		cfg.Debug = true
		cfg.NoLog = false
	case liblog.NilLevel:
		cfg.Debug = false
		cfg.NoLog = true
	default:
		cfg.Debug = false
		cfg.NoLog = false
	}

	// Limits

	if c.Limits.MaxConn > 0 {
		cfg.MaxConn = c.Limits.MaxConn
	}

	if c.Limits.MaxSubs > 0 {
		cfg.MaxSubs = c.Limits.MaxSubs
	}

	if c.Limits.PingInterval > 0 {
		cfg.PingInterval = c.Limits.PingInterval
	}

	if c.Limits.MaxPingsOut > 0 {
		cfg.MaxPingsOut = c.Limits.MaxPingsOut
	}

	if c.Limits.MaxControlLine > 0 {
		cfg.MaxControlLine = c.Limits.MaxControlLine
	}

	if c.Limits.MaxPayload > 0 {
		cfg.MaxPayload = c.Limits.MaxPayload
	}

	if c.Limits.MaxPending > 0 {
		cfg.MaxPending = c.Limits.MaxPending
	}

	if c.Limits.WriteDeadline > 0 {
		cfg.WriteDeadline = c.Limits.WriteDeadline
	}

	if c.Limits.RQSubsSweep > 0 {
		cfg.RQSubsSweep = c.Limits.RQSubsSweep
	}

	if c.Limits.MaxClosedClients > 0 {
		cfg.MaxClosedClients = c.Limits.MaxClosedClients
	}

	if c.Limits.LameDuckDuration > 0 {
		cfg.LameDuckDuration = c.Limits.LameDuckDuration
	}

	// Cluster

	if c.Cluster.Host != "" {
		cfg.Cluster.Host = c.Cluster.Host
	}

	if c.Cluster.Port > 0 {
		cfg.Cluster.Port = c.Cluster.Port
	}

	if c.Cluster.ListenStr != "" {
		cfg.Cluster.ListenStr = c.Cluster.ListenStr
	}

	if c.Cluster.Advertise != "" {
		cfg.Cluster.Advertise = c.Cluster.Advertise
	}

	if c.Cluster.NoAdvertise {
		cfg.Cluster.NoAdvertise = true
	}

	if c.Cluster.ConnectRetries > 0 {
		cfg.Cluster.ConnectRetries = c.Cluster.ConnectRetries
	}

	if c.Cluster.Username != "" {
		cfg.Cluster.Username = c.Cluster.Username
	}

	if c.Cluster.Password != "" {
		cfg.Cluster.Password = c.Cluster.Password
	}

	if c.Cluster.AuthTimeout > 0 {
		cfg.Cluster.AuthTimeout = float64(c.Cluster.AuthTimeout) / float64(time.Second)
	}

	if cfg.Cluster.Permissions == nil {
		cfg.Cluster.Permissions = &natsrv.RoutePermissions{
			Import: &natsrv.SubjectPermission{
				Allow: make([]string, 0),
				Deny:  make([]string, 0),
			},
			Export: &natsrv.SubjectPermission{
				Allow: make([]string, 0),
				Deny:  make([]string, 0),
			},
		}
	}

	if len(c.Cluster.Permissions.Import.Allow) > 0 {
		for _, r := range c.Cluster.Permissions.Import.Allow {
			if r != "" {
				cfg.Cluster.Permissions.Import.Allow = append(cfg.Cluster.Permissions.Import.Allow, r)
			}
		}
	}

	if len(c.Cluster.Permissions.Import.Deny) > 0 {
		for _, r := range c.Cluster.Permissions.Import.Deny {
			if r != "" {
				cfg.Cluster.Permissions.Import.Deny = append(cfg.Cluster.Permissions.Import.Deny, r)
			}
		}
	}

	if len(c.Cluster.Permissions.Export.Allow) > 0 {
		for _, r := range c.Cluster.Permissions.Export.Allow {
			if r != "" {
				cfg.Cluster.Permissions.Export.Allow = append(cfg.Cluster.Permissions.Export.Allow, r)
			}
		}
	}

	if len(c.Cluster.Permissions.Export.Deny) > 0 {
		for _, r := range c.Cluster.Permissions.Export.Deny {
			if r != "" {
				cfg.Cluster.Permissions.Export.Deny = append(cfg.Cluster.Permissions.Export.Deny, r)
			}
		}
	}

	if c.Cluster.TLS {
		if t, e := c.Cluster.TLSConfig.NewFrom(defaultTls); e != nil {
			return nil, e
		} else {
			cfg.Cluster.TLSConfig = t.TlsConfig("")
		}
		if c.Cluster.TLSTimeout > 0 {
			cfg.Cluster.TLSTimeout = float64(c.Cluster.TLSTimeout) / float64(time.Second)
		}
	} else {
		cfg.Cluster.TLSConfig = nil
		cfg.Cluster.TLSTimeout = 0
	}

	// Auth

	if c.Auth.AuthTimeout > 0 {
		cfg.AuthTimeout = float64(c.Auth.AuthTimeout) / float64(time.Second)
	}

	if len(cfg.Users) == 0 {
		cfg.Users = make([]*natsrv.User, 0)
	}

	if len(c.Auth.Users) > 0 {
		for _, u := range c.Auth.Users {
			if u.Username == "" {
				continue
			}
			if u.Password == "" {
				continue
			}

			usr := &natsrv.User{
				Username: u.Username,
				Password: u.Password,
				Permissions: &natsrv.Permissions{
					Publish: &natsrv.SubjectPermission{
						Allow: make([]string, 0),
						Deny:  make([]string, 0),
					},
					Subscribe: &natsrv.SubjectPermission{
						Allow: make([]string, 0),
						Deny:  make([]string, 0),
					},
				},
			}

			if len(u.Permissions.Publish.Allow) > 0 {
				for _, r := range u.Permissions.Publish.Allow {
					if r != "" {
						usr.Permissions.Publish.Allow = append(usr.Permissions.Publish.Allow, r)
					}
				}
			}

			if len(u.Permissions.Publish.Deny) > 0 {
				for _, r := range u.Permissions.Publish.Deny {
					if r != "" {
						usr.Permissions.Publish.Deny = append(usr.Permissions.Publish.Deny, r)
					}
				}
			}

			if len(u.Permissions.Subscribe.Allow) > 0 {
				for _, r := range u.Permissions.Subscribe.Allow {
					if r != "" {
						usr.Permissions.Subscribe.Allow = append(usr.Permissions.Subscribe.Allow, r)
					}
				}
			}

			if len(u.Permissions.Subscribe.Deny) > 0 {
				for _, r := range u.Permissions.Subscribe.Deny {
					if r != "" {
						usr.Permissions.Subscribe.Deny = append(usr.Permissions.Subscribe.Deny, r)
					}
				}
			}

			cfg.Users = append(cfg.Users, usr)
		}
	}

	return cfg, nil
}
