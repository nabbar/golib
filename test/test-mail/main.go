//go:build examples
// +build examples

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
	libpool "github.com/nabbar/golib/mailPooler"
	libtpl "github.com/nabbar/golib/mailer"
	libsem "github.com/nabbar/golib/semaphore"
	libsmtp "github.com/nabbar/golib/smtp"
)

const (
	_ConfigSmtpDSN   = "login@email-example.com:password@tcp4(smtp.mail.example.com:25)/starttls?ServerName=mail.domain.com"
	_ConfigEmailFrom = "email@example.com"
	_ConfigEmailTo   = "email@example.com"
	_ConfigSubject   = "Testing Send Mail"
)

var (
	ctx context.Context
	cnl context.CancelFunc
)

func init() {
	liberr.SetModeReturnError(liberr.ErrorReturnCodeErrorTraceFull)

	ctx, cnl = context.WithCancel(context.TODO())

	liblog.SetLevel(liblog.DebugLevel)
	if err := liblog.GetDefault().SetOptions(&liblog.Options{
		DisableStandard:  false,
		DisableStack:     false,
		DisableTimestamp: false,
		EnableTrace:      true,
		TraceFilter:      "",
		DisableColor:     false,
	}); err != nil {
		panic(err)
	}
}

func main() {
	var (
		cli libpool.Pooler
		err liberr.Error
		sem = libsem.NewSemaphoreWithContext(ctx, 0)
	)

	defer func() {
		sem.DeferMain()
		cnl()
		if cli != nil {
			cli.Close()
		}
	}()

	//cli = getSmtp()
	cli = getPool()

	snd := getSendMail(getTemplate())

	//	err = getSendMail(getTemplate()).SendClose(ctx, cli)
	for i := 0; i < 5; i++ {
		_ = sem.NewWorker()
		go func() {
			defer sem.DeferWorker()
			err = snd.Send(ctx, cli)
			liblog.FatalLevel.LogErrorCtxf(liblog.InfoLevel, "[sender] sending email", err)
			time.Sleep(2 * time.Second)
		}()
	}

	_ = sem.WaitAll()
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
	cfg, err := libsmtp.NewConfig(_ConfigSmtpDSN)
	liblog.FatalLevel.LogErrorCtxf(liblog.InfoLevel, "[smtp] config parsing", err)

	/* #nosec */
	//nolint #nosec
	s, err := libsmtp.NewSMTP(cfg, &tls.Config{})
	liblog.FatalLevel.LogErrorCtxf(liblog.InfoLevel, "[smtp] init", err)
	liblog.FatalLevel.LogErrorCtxf(liblog.InfoLevel, "[smtp] checking working", s.Check(ctx))

	return s
}

func getPool() libpool.Pooler {
	cfg := &libpool.Config{
		Max:  2,
		Wait: 10 * time.Second,
	}
	cfg.SetFuncCaller(func() liberr.Error {
		liblog.FatalLevel.LogErrorCtxf(liblog.InfoLevel, "[mail pooler] reset counter", nil)
		return nil
	})

	p := libpool.New(cfg, getSmtp())
	liblog.FatalLevel.LogErrorCtxf(liblog.InfoLevel, "[mail pooler] init", nil)

	return p
}

func getSendMail(tpl libtpl.Mailer) libsnd.Sender {
	m := libsnd.New()

	m.Email().SetFrom(_ConfigEmailFrom)
	m.Email().SetRecipients(libsnd.RecipientTo, _ConfigEmailTo)

	m.SetCharset("UTF-8")
	m.SetPriority(libsnd.PriorityNormal)
	m.SetSubject(_ConfigSubject)
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
