package httpcli

import (
	"fmt"
	"net/http"
	"time"

	libval "github.com/go-playground/validator/v10"

	libtls "github.com/nabbar/golib/certificates"
	liberr "github.com/nabbar/golib/errors"
)

type OptionForceIP struct {
	Enable bool   `json:"enable" yaml:"enable" toml:"enable" mapstructure:"enable"`
	Net    string `json:"net,omitempty" yaml:"net,omitempty" toml:"net,omitempty" mapstructure:"net,omitempty"`
	IP     string `json:"ip,omitempty" yaml:"ip,omitempty" toml:"ip,omitempty" mapstructure:"ip,omitempty"`
}

type OptionTLS struct {
	Enable bool          `json:"enable" yaml:"enable" toml:"enable" mapstructure:"enable"`
	Config libtls.Config `json:"tls" yaml:"tls" toml:"tls" mapstructure:"tls"`
}

type Options struct {
	Timeout time.Duration `json:"timeout" yaml:"timeout" toml:"timeout" mapstructure:"timeout"`
	Http2   bool          `json:"http2" yaml:"http2" toml:"http2" mapstructure:"http2"`
	TLS     OptionTLS     `json:"tls" yaml:"tls" toml:"tls" mapstructure:"tls"`
	ForceIP OptionForceIP `json:"force_ip" yaml:"force_ip" toml:"force_ip" mapstructure:"force_ip"`
}

func (o Options) Validate() liberr.Error {
	var e = ErrorValidatorError.Error(nil)

	if err := libval.New().Struct(o); err != nil {
		if er, ok := err.(*libval.InvalidValidationError); ok {
			e.AddParent(er)
		}

		for _, er := range err.(libval.ValidationErrors) {
			//nolint #goerr113
			e.AddParent(fmt.Errorf("config field '%s' is not validated by constraint '%s'", er.Namespace(), er.ActualTag()))
		}
	}

	if !e.HasParent() {
		e = nil
	}

	return e
}

func (o Options) GetClient(def libtls.TLSConfig, servername string) (*http.Client, liberr.Error) {
	var tls libtls.TLSConfig

	if t, e := o._GetTLS(def); e != nil {
		return nil, e
	} else {
		tls = t
	}

	if o.ForceIP.Enable {
		return GetClientTlsForceIp(GetNetworkFromString(o.ForceIP.Net), o.ForceIP.IP, servername, tls, o.Http2, o.Timeout)
	} else {
		return GetClientTls(servername, tls, o.Http2, o.Timeout)
	}
}

func (o Options) _GetTLS(def libtls.TLSConfig) (libtls.TLSConfig, liberr.Error) {
	if o.TLS.Enable {
		return o.TLS.Config.NewFrom(def)
	} else {
		return libtls.Default.Clone(), nil
	}
}
