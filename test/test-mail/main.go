/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package main

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/matcornic/hermes/v2"
	liberr "github.com/nabbar/golib/errors"
	libiot "github.com/nabbar/golib/ioutils"
	liblog "github.com/nabbar/golib/logger"
	libsnd "github.com/nabbar/golib/mail"
	libtpl "github.com/nabbar/golib/mailer"
	libsmtp "github.com/nabbar/golib/smtp"
)

const (
	CONFIG_SMTP_DSN   = "login@email-example.com:password@tcp4(smtp.mail.example.com:25)/starttls?ServerName=mail.domain.com"
	CONFIG_EMAIL_FROM = "email@example.com"
	CONFIG_EMAIL_TO   = "email@example.com"
	CONFIG_SUBJECT    = "Testing Send Mail"
)

var (
	ctx context.Context
	cnl context.CancelFunc
)

func init() {
	liblog.EnableColor()
	liblog.AddGID(true)
	liblog.FileTrace(true)
	liblog.SetFormat(liblog.TextFormat)
	liblog.SetLevel(liblog.DebugLevel)
	liberr.SetModeReturnError(liberr.ErrorReturnCodeErrorTraceFull)

	ctx, cnl = context.WithCancel(context.TODO())
}

func main() {
	var (
		cli libsmtp.SMTP
		err liberr.Error
	)

	defer func() {
		cnl()
		if cli != nil {
			cli.Close()
		}
	}()

	cli = getSmtp()

	err = getSendMail(getTemplate()).SendClose(ctx, cli)

	liblog.FatalLevel.LogErrorCtxf(liblog.InfoLevel, "sending email", err)
}

func getTemplate() libtpl.Mailer {
	body := &hermes.Body{
		Name: "Jon Snow",
		Intros: []string{
			"Welcome to Test! We're very excited to have you on board.",
		},
		Actions: []hermes.Action{
			{
				Instructions: "To get started with Hermes, please click here:",
				Button: hermes.Button{
					Color: "#22BC66", // Optional action button color
					Text:  "Confirm your account",
					Link:  "https://example.com/confirm?token=123456789abcdef123456789abcdef",
				},
			},
		},
		Outros: []string{
			"Need help, or have questions? Just reply to this email, we'd love to help.",
		},
	}

	tpl := libtpl.New()
	tpl.SetTextDirection(libtpl.LeftToRight)
	tpl.SetTheme(libtpl.ThemeDefault)
	tpl.SetCSSInline(false)

	tpl.SetTroubleText("If you’re having trouble with the button '{ACTION}', copy and paste the URL below into your web browser.")
	tpl.SetCopyright("Copyright © 2021 Nabbar. All rights reserved.")
	tpl.SetLogo("https://example.com/logo.png")
	tpl.SetLink("https://example.com/")
	tpl.SetName("Nabbar Test Mail")

	tpl.SetBody(body)

	return tpl
}

func getSmtp() libsmtp.SMTP {
	cfg, err := libsmtp.NewConfig(CONFIG_SMTP_DSN)
	liblog.FatalLevel.LogErrorCtxf(liblog.InfoLevel, "smtp config parsing", err)

	s, err := libsmtp.NewSMTP(cfg, &tls.Config{})
	liblog.FatalLevel.LogErrorCtxf(liblog.InfoLevel, "smtp create client", err)
	liblog.FatalLevel.LogErrorCtxf(liblog.InfoLevel, "smtp checking working", s.Check(ctx))

	return s
}

func getSendMail(tpl libtpl.Mailer) libsnd.Sender {
	m := libsnd.New()

	m.Email().SetFrom(CONFIG_EMAIL_FROM)
	m.Email().SetRecipients(libsnd.RecipientTo, CONFIG_EMAIL_TO)

	m.SetCharset("UTF-8")
	m.SetPriority(libsnd.PriorityNormal)
	m.SetSubject(CONFIG_SUBJECT)
	m.SetEncoding(libsnd.EncodingBinary)
	m.SetDateTime(time.Now())

	bh, err := tpl.GenerateHTML()
	liblog.FatalLevel.LogErrorCtxf(liblog.InfoLevel, "template generating html", err)
	m.SetBody(libsnd.ContentHTML, libiot.NewBufferReadCloser(bh))

	bt, err := tpl.GeneratePlainText()
	liblog.FatalLevel.LogErrorCtxf(liblog.InfoLevel, "template generating text", err)
	m.AddBody(libsnd.ContentPlainText, libiot.NewBufferReadCloser(bt))

	s, e := m.Sender()
	liblog.FatalLevel.LogErrorCtxf(liblog.InfoLevel, "mail create sender", e)

	return s
}
