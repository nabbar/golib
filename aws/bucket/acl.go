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

package bucket

import (
	adkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdkstp "github.com/aws/aws-sdk-go-v2/service/s3/types"
	libhlp "github.com/nabbar/golib/aws/helper"
	liberr "github.com/nabbar/golib/errors"
)

type ACLHeader uint8

const (
	ACLHeaderNone ACLHeader = iota
	ACLHeaderFullControl
	ACLHeaderWrite
	ACLHeaderRead
	ACLHeaderWriteACP
	ACLHeaderReadACP
)

type ACLHeaders map[ACLHeader]string

// for GetACL
// see : https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetBucketAcl.html

func (cli *client) GetACL() (*sdkstp.AccessControlPolicy, liberr.Error) {
	out, err := cli.s3.GetBucketAcl(cli.GetContext(), &sdksss.GetBucketAclInput{
		Bucket: cli.GetBucketAws(),
	})

	res := &sdkstp.AccessControlPolicy{
		Owner: &sdkstp.Owner{
			DisplayName: nil,
			ID:          nil,
		},
		Grants: make([]sdkstp.Grant, 0),
	}

	if err != nil {
		return nil, cli.GetError(err)
	} else if out == nil {
		return nil, libhlp.ErrorResponse.Error(nil)
	} else if out.Owner == nil || out.Grants == nil || len(out.Grants) < 1 {
		return res, nil
	}

	res.Owner = out.Owner
	res.Grants = out.Grants

	// MarshalValue always return error as nil
	return res, nil
}

// for SetACL
//example value : emailAddress="xyz@amazon.com"
//example value : uri="http://acs.amazonaws.com/groups/global/AllUsers"
//example value : uri="http://acs.amazonaws.com/groups/s3/LogDelivery", emailAddress="xyz@amazon.com"
// for more info, see : https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketAcl.html#API_PutBucketAcl_RequestSyntax

func (cli *client) SetACL(ACP *sdkstp.AccessControlPolicy, cannedACL sdkstp.BucketCannedACL, header ACLHeaders) liberr.Error {
	in := &sdksss.PutBucketAclInput{
		Bucket: cli.GetBucketAws(),
	}

	return cli.setACLInput(in, ACP, cannedACL, header)
}

func (cli *client) SetACLPolicy(ACP *sdkstp.AccessControlPolicy) liberr.Error {
	in := &sdksss.PutBucketAclInput{
		Bucket: cli.GetBucketAws(),
	}

	return cli.setACLInput(in, ACP, "", nil)
}

func (cli *client) SetACLHeader(cannedACL sdkstp.BucketCannedACL, header ACLHeaders) liberr.Error {
	in := &sdksss.PutBucketAclInput{
		Bucket: cli.GetBucketAws(),
	}

	return cli.setACLInput(in, nil, cannedACL, header)
}

func (cli *client) setACLInput(in *sdksss.PutBucketAclInput, ACP *sdkstp.AccessControlPolicy, cannedACL sdkstp.BucketCannedACL, header ACLHeaders) liberr.Error {
	if ACP != nil {
		in.AccessControlPolicy = ACP
	}

	if cannedACL != "" {
		in.ACL = cannedACL
	}

	if header != nil {
		for k, v := range header {
			switch k {
			case ACLHeaderFullControl:
				in.GrantFullControl = adkaws.String(v)
			case ACLHeaderRead:
				in.GrantRead = adkaws.String(v)
			case ACLHeaderWrite:
				in.GrantWrite = adkaws.String(v)
			case ACLHeaderReadACP:
				in.GrantReadACP = adkaws.String(v)
			case ACLHeaderWriteACP:
				in.GrantWriteACP = adkaws.String(v)
			}
		}
	}

	_, err := cli.s3.PutBucketAcl(cli.GetContext(), in)
	return cli.GetError(err)
}
