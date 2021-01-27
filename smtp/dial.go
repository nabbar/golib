package smtp

import (
	"context"
	"crypto/tls"
	"net"
	"net/smtp"

	liberr "github.com/nabbar/golib/errors"
)

func (s *smtpClient) dialTLS(ctx context.Context, addr string, tlsConfig *tls.Config) (con net.Conn, err liberr.Error) {
	defer func() {
		if err != nil && con != nil {
			_ = con.Close()
		}
	}()

	if con, err = s.dial(ctx, addr); err != nil {
		return
	}

	if con = tls.Client(con, tlsConfig); con == nil {
		err = ErrorSMTPClientStartTLS.Error(nil)
	}

	return
}

func (s *smtpClient) dial(ctx context.Context, addr string) (con net.Conn, err liberr.Error) {
	var e error

	defer func() {
		if err != nil && con != nil {
			_ = con.Close()
		}
	}()

	d := net.Dialer{}

	if con, e = d.DialContext(ctx, s.cfg.GetNet().string(), addr); e != nil {
		return con, ErrorSMTPDial.ErrorParent(e)
	}

	return
}

func (s *smtpClient) dialStarttls(ctx context.Context, addr string, tlsConfig *tls.Config) (con net.Conn, err liberr.Error) {
	defer func() {
		if err != nil && con != nil {
			_ = con.Close()
		}
	}()

	con, err = s.dial(ctx, addr)
	return
}

func (s *smtpClient) client(ctx context.Context, addr string, tlsConfig *tls.Config) (cli *smtp.Client, con net.Conn, err liberr.Error) {
	var e error

	defer func() {
		if err != nil {
			if cli != nil {
				_ = cli.Close()
			}
			if con != nil {
				_ = con.Close()
			}
		}
	}()

	if s.cfg.GetTls() == TLS_STARTTLS && tlsConfig == nil {
		err = ErrorParamsEmpty.Error(nil)
		return
	} else if s.cfg.GetTls() == TLS_TLS && tlsConfig == nil {
		err = ErrorParamsEmpty.Error(nil)
		return
	}

	if s.cfg.GetTls() == TLS_TLS && tlsConfig != nil {
		if con, err = s.dialTLS(ctx, addr, tlsConfig); err != nil {
			return
		} else if cli, e = smtp.NewClient(con, addr); e != nil {
			err = ErrorSMTPClientInit.ErrorParent(e)
			return
		}
	} else {
		if con, err = s.dial(ctx, addr); err != nil {
			return
		} else if cli, e = smtp.NewClient(con, addr); e != nil {
			err = ErrorSMTPClientInit.ErrorParent(e)
			return
		}

		try := s.checkExtension(cli, "STARTTLS")

		if s.cfg.GetTls() == TLS_STARTTLS || try {
			if e = cli.StartTLS(tlsConfig); e != nil && !try {
				err = ErrorSMTPClientStartTLS.ErrorParent(e)
				return
			} else if e == nil && try {
				s.cfg.SetTls(TLS_STARTTLS)
			}
		}
	}

	return
}

func (s *smtpClient) tryClient(ctx context.Context, addr string, tlsConfig *tls.Config) (cli *smtp.Client, con net.Conn, err liberr.Error) {
	cli, con, err = s.client(ctx, addr, tlsConfig)

	if err == nil {
		return
	}

	switch s.cfg.GetTls() {
	case TLS_TLS:
		s.cfg.SetTls(TLS_STARTTLS)
		return s.tryClient(ctx, addr, tlsConfig)
	case TLS_STARTTLS, TLS_NONE:
		return
	}

	return
}

func (s *smtpClient) auth(cli *smtp.Client, addr string) liberr.Error {
	usr := s.cfg.GetUser()
	pwd := s.cfg.GetPass()
	err := ErrorSMTPClientAuth.Error(nil)

	if usr == "" && pwd == "" {
		return nil
	}
	/*
		if e := cli.Auth(smtp.CRAMMD5Auth(usr, pwd)); e != nil {
			err.AddParent(e)
		} else {
			return nil
		}
	*/
	if e := cli.Auth(smtp.PlainAuth("", usr, pwd, addr)); e != nil {
		err.AddParent(e)
	} else {
		return nil
	}

	return err
}

func (s *smtpClient) checkExtension(cli *smtp.Client, ext string) bool {
	ok, _ := cli.Extension(ext)
	return ok
}
