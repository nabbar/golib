package njs_smtp

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"os"

	"github.com/olekukonko/tablewriter"

	"github.com/jaytaylor/html2text"
)

type mailTemplate struct {
	data interface{}
	char string
	opt  html2text.Options
	tpl  *template.Template
}

func (m *mailTemplate) SetCharset(char string) {
	m.char = char
}

func (m mailTemplate) GetCharset() string {
	return m.char
}

func (m *mailTemplate) SetTextOption(opt html2text.Options) {
	m.opt = opt
}

func (m mailTemplate) GetTextOption() html2text.Options {
	return m.opt
}

func (m mailTemplate) GetBufferHtml(data interface{}) (*bytes.Buffer, error) {
	var res = bytes.NewBuffer(make([]byte, 0))

	if data == nil {
		data = m.data
	}

	if err := m.tpl.Execute(res, data); err != nil {
		return nil, err
	}

	return res, nil
}

func (m mailTemplate) GetBufferText(data interface{}) (*bytes.Buffer, error) {
	var (
		res = bytes.NewBuffer(make([]byte, 0))
		str string
	)

	if buf, err := m.GetBufferHtml(data); err != nil {
		return nil, err
	} else if str, err = html2text.FromReader(buf, m.opt); err != nil {
		return nil, err
	} else if _, err = res.WriteString(str); err != nil {
		return nil, err
	}

	return res, nil
}

func (m mailTemplate) GetBufferRich(data interface{}) (*bytes.Buffer, error) {
	panic("implement me")
}

func (m *mailTemplate) RegisterData(data interface{}) {
	m.data = data
}

func (m mailTemplate) IsEmpty() bool {
	if m.tpl == nil {
		return true
	}

	if m.tpl.DefinedTemplates() == "" {
		return true
	}

	return false
}

func (m mailTemplate) Clone() (MailTemplate, error) {
	res := &mailTemplate{
		data: nil,
		char: m.char,
		opt:  m.opt,
		tpl:  nil,
	}

	if tpl, err := m.tpl.Clone(); err != nil {
		return nil, err
	} else {
		res.tpl = tpl
	}

	return res, nil
}

type MailTemplate interface {
	Clone() (MailTemplate, error)

	IsEmpty() bool

	SetCharset(char string)
	GetCharset() string

	SetTextOption(opt html2text.Options)
	GetTextOption() html2text.Options

	GetBufferHtml(data interface{}) (*bytes.Buffer, error)
	GetBufferText(data interface{}) (*bytes.Buffer, error)
	GetBufferRich(data interface{}) (*bytes.Buffer, error)

	RegisterData(data interface{})
}

func NewMailTemplate(name, tpl string, isFile bool) (MailTemplate, error) {
	var (
		err error
		res = &mailTemplate{
			data: nil,
			tpl:  template.New(name),
			opt: html2text.Options{
				PrettyTables: false,
				PrettyTablesOptions: &html2text.PrettyTablesOptions{
					AutoFormatHeader:     false,
					AutoWrapText:         false,
					ReflowDuringAutoWrap: false,
					ColWidth:             0,
					ColumnSeparator:      "",
					RowSeparator:         "",
					CenterSeparator:      "",
					HeaderAlignment:      0,
					FooterAlignment:      0,
					Alignment:            0,
					ColumnAlignment:      nil,
					NewLine:              "",
					HeaderLine:           false,
					RowLine:              false,
					AutoMergeCells:       false,
					Borders:              tablewriter.Border{},
				},
				OmitLinks: false,
			},
		}
	)

	if isFile {
		var fs []byte
		if _, err = os.Stat(tpl); err != nil {
			return nil, err
		} else if fs, err = ioutil.ReadFile(tpl); err != nil {
			return nil, err
		}

		tpl = string(fs)
	}

	if res.tpl, err = res.tpl.Parse(tpl); err != nil {
		return nil, err
	}

	return res, err
}
