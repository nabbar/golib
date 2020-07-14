package smtp

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"os"

	"github.com/jaytaylor/html2text"
	"github.com/olekukonko/tablewriter"

	. "github.com/nabbar/golib/errors"
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

func (m mailTemplate) GetBufferHtml(data interface{}) (*bytes.Buffer, Error) {
	var res = bytes.NewBuffer(make([]byte, 0))

	if data == nil {
		data = m.data
	}

	if err := m.tpl.Execute(res, data); err != nil {
		return nil, TEMPLATE_EXECUTE.ErrorParent(err)
	}

	return res, nil
}

func (m mailTemplate) GetBufferText(data interface{}) (*bytes.Buffer, Error) {
	var (
		res = bytes.NewBuffer(make([]byte, 0))
		str string
		e   error
	)

	if buf, err := m.GetBufferHtml(data); err != nil {
		return nil, err
	} else if str, e = html2text.FromReader(buf, m.opt); e != nil {
		return nil, TEMPLATE_HTML2TEXT.ErrorParent(e)
	} else if _, e = res.WriteString(str); e != nil {
		return nil, BUFFER_WRITE_STRING.ErrorParent(e)
	}

	return res, nil
}

func (m mailTemplate) GetBufferRich(data interface{}) (*bytes.Buffer, Error) {
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

func (m mailTemplate) Clone() (MailTemplate, Error) {
	res := &mailTemplate{
		data: nil,
		char: m.char,
		opt:  m.opt,
		tpl:  nil,
	}

	if tpl, err := m.tpl.Clone(); err != nil {
		return nil, TEMPLATE_CLONE.ErrorParent(err)
	} else {
		res.tpl = tpl
	}

	return res, nil
}

type MailTemplate interface {
	Clone() (MailTemplate, Error)

	IsEmpty() bool

	SetCharset(char string)
	GetCharset() string

	SetTextOption(opt html2text.Options)
	GetTextOption() html2text.Options

	GetBufferHtml(data interface{}) (*bytes.Buffer, Error)
	GetBufferText(data interface{}) (*bytes.Buffer, Error)
	GetBufferRich(data interface{}) (*bytes.Buffer, Error)

	RegisterData(data interface{})
}

func NewMailTemplate(name, tpl string, isFile bool) (MailTemplate, Error) {
	var (
		err error
		res = &mailTemplate{
			data: nil,
			tpl:  template.New(name),
			opt: html2text.Options{
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
			},
		}
	)

	if isFile {
		var fs []byte
		// #nosec
		if _, err = os.Stat(tpl); err != nil {
			return nil, FILE_STAT.ErrorParent(err)
		} else if fs, err = ioutil.ReadFile(tpl); err != nil {
			return nil, FILE_READ.ErrorParent(err)
		}

		tpl = string(fs)
	}

	if res.tpl, err = res.tpl.Parse(tpl); err != nil {
		return nil, TEMPLATE_PARSING.ErrorParent(err)
	}

	return res, nil
}
