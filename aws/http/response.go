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

package http

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime"
	"net/http"
)

var (
	ErrInvalidResponse = fmt.Errorf("invalid response")
)

type Config interface {
	GetRegion() string
	GetAccessKey() string
	GetSecretKey() string
}

type ErrorResponse struct {
	XMLName   xml.Name `xml:"Error"`
	Code      string   `xml:"Code" json:"code"`
	Message   string   `xml:"Message" json:"message"`
	RequestID string   `xml:"RequestId" json:"requestId"`
}

func (e ErrorResponse) Error() string {
	return fmt.Sprintf("request ID %s occure an aws response error %s: %s", e.RequestID, e.Code, e.Message)
}

type ErrorStatus struct {
	Status  string `xml:"Code" json:"code"`
	Message string `xml:"Message" json:"message"`
}

func (e ErrorStatus) Error() string {
	return fmt.Sprintf("invalid response status code (%s): %s", e.Status, e.Message)
}

func Response(rsp *http.Response, model any) error {
	defer func() {
		if rsp != nil && rsp.Body != nil {
			_ = rsp.Body.Close()
		}
	}()

	var (
		err error
		buf = bytes.NewBuffer(make([]byte, 0))
		cnj = mime.TypeByExtension(".json")
		cnx = mime.TypeByExtension(".xml")
	)

	if rsp == nil {
		return ErrInvalidResponse
	} else if rsp.Body != nil {
		if _, e := io.Copy(buf, rsp.Body); e != nil {
			return e
		}
	}

	if tp := rsp.Header.Get("Content-Type"); tp == cnj {
		err = responseJson(buf, model)
	} else if tp != cnx {
		err = responseXml(buf, model)
	} else {
		return ErrInvalidResponse
	}

	if rsp.StatusCode < 200 || rsp.StatusCode >= 300 {
		if err != nil {
			return err
		} else {
			return &ErrorStatus{
				Status:  rsp.Status,
				Message: truncateBuf(buf),
			}
		}
	} else if err != nil {
		return err
	}

	return nil
}

func truncateBuf(buf *bytes.Buffer) string {
	if buf.Len() > 255 {
		return buf.String()[:255]
	} else {
		return buf.String()
	}
}

func responseJson(buf *bytes.Buffer, model any) error {
	if e := json.Unmarshal(buf.Bytes(), model); e != nil {
		return responseJsonError(buf)
	}

	return nil
}

func responseJsonError(buf *bytes.Buffer) error {
	var err = ErrorResponse{}

	if e := json.Unmarshal(buf.Bytes(), &err); e != nil {
		return e
	} else {
		return err
	}
}

func responseXml(buf *bytes.Buffer, model any) error {
	if e := xml.Unmarshal(buf.Bytes(), model); e != nil {
		return responseXmlError(buf)
	}

	return nil
}

func responseXmlError(buf *bytes.Buffer) error {
	var err = ErrorResponse{}

	if e := xml.Unmarshal(buf.Bytes(), &err); e != nil {
		return e
	} else {
		return err
	}
}
