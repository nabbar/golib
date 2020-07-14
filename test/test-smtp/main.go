package main

import (
	"bytes"
	"crypto/tls"
	"fmt"

	"github.com/jaytaylor/html2text"
	"github.com/olekukonko/tablewriter"

	"github.com/nabbar/golib/logger"
	"github.com/nabbar/golib/smtp"
)

const (
	CONFIG_SMTP_DSN      = "user@example.com:test_password@tcp4(mail.example.com:25)/starttls?ServerName=mail.example.com"
	CONFIG_EMAIL_FROM    = "sender@example.com"
	CONFIG_EMAIL_TO      = "recipient@example.com"
	CONFIG_MAILER        = "Nabbar SMTP Tester"
	CONFIG_SUBJECT       = "Testing Send Mail"
	CONFIG_TESTMODE      = false
	CONFIG_TEMPLATE_TEST = `<html><head></head><body><b>Hello {{.Name}}</b>, this is a test e-mail sent by <i>Go</i> with package nabbar/golib/smtp.</body></html>`
)

func main() {
	var (
		cfg  smtp.SMTP
		tpl  smtp.MailTemplate
		snd  smtp.SendMail
		err  error
		buff = bytes.NewBuffer(make([]byte, 0))
	)

	logger.EnableColor()
	logger.AddGID(true)
	logger.FileTrace(true)
	logger.SetFormat(logger.TextFormat)
	logger.SetLevel(logger.DebugLevel)

	tpl, err = smtp.NewMailTemplate("mail", CONFIG_TEMPLATE_TEST, false)
	logger.FatalLevel.LogErrorCtxf(logger.InfoLevel, "mail template parsing", err)
	tpl.SetCharset("utf-8")
	tpl.RegisterData(struct {
		Name string
	}{Name: "éloïse"})
	tpl.SetTextOption(html2text.Options{
		PrettyTables: true,
		PrettyTablesOptions: &html2text.PrettyTablesOptions{
			AutoFormatHeader:     true,
			AutoWrapText:         true,
			ReflowDuringAutoWrap: true,
			ColWidth:             tablewriter.MAX_ROW_WIDTH,
			ColumnSeparator:      tablewriter.COLUMN,
			RowSeparator:         tablewriter.ROW,
			CenterSeparator:      tablewriter.CENTER,
			HeaderAlignment:      tablewriter.ALIGN_DEFAULT,
			FooterAlignment:      tablewriter.ALIGN_DEFAULT,
			Alignment:            tablewriter.ALIGN_DEFAULT,
			ColumnAlignment:      []int{},
			NewLine:              tablewriter.NEWLINE,
			HeaderLine:           true,
			RowLine:              false,
			AutoMergeCells:       false,
			Borders:              tablewriter.Border{Left: true, Right: true, Bottom: true, Top: true},
		},
		OmitLinks: true,
	})

	if p, e := tpl.GetBufferHtml(nil); e == nil {
		fmt.Printf("\n\n\n\t >> HTML Mail : \n")
		print(p.String())
		fmt.Printf("\n\n")
	}

	if p, e := tpl.GetBufferText(nil); e == nil {
		fmt.Printf("\n\n\n\t >> Text Mail : \n")
		print(p.String())
		fmt.Printf("\n\n")
	}

	snd = smtp.NewSendMail()
	snd.SetTo(smtp.MailAddressParser(CONFIG_EMAIL_TO))
	snd.SetFrom(smtp.MailAddressParser(CONFIG_EMAIL_FROM))
	snd.SetMailer(CONFIG_MAILER)
	snd.SetSubject(CONFIG_SUBJECT)
	snd.SetTestMode(CONFIG_TESTMODE)
	snd.SetForceOnly(smtp.CONTENTTYPE_HTML)
	snd.SetHtml(tpl)

	cfg, err = smtp.NewSMTP(CONFIG_SMTP_DSN, &tls.Config{})
	logger.FatalLevel.LogErrorCtxf(logger.InfoLevel, "smtp config parsing", err)
	logger.FatalLevel.LogErrorCtxf(logger.InfoLevel, "smtp checking working", cfg.Check())

	err, buff = snd.SendSMTP(cfg)
	logger.FatalLevel.LogErrorCtxf(logger.InfoLevel, "Sending Mail", err)

	fmt.Printf("\n\n\n\t >> Buff Mail : \n")
	print(buff.String())
	fmt.Printf("\n\n")

}
