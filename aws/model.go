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

package aws

import (
	"context"

	"github.com/nabbar/golib/aws/bucket"
	"github.com/nabbar/golib/aws/group"
	"github.com/nabbar/golib/aws/object"
	"github.com/nabbar/golib/aws/policy"
	"github.com/nabbar/golib/aws/role"
	"github.com/nabbar/golib/aws/user"
	"github.com/nabbar/golib/errors"
)

func (c *client) ForcePathStyle(ctx context.Context, enabled bool) errors.Error {
	c.p = enabled

	if s, e := c.newClientS3(ctx, nil); e != nil {
		return e
	} else {
		c.s = s
	}

	return nil
}

func (c *client) Config() Config {
	return c.c
}

func (c *client) Bucket() bucket.Bucket {
	return bucket.New(c.x, c.c.GetBucketName(), c.c.GetRegion(), c.i, c.s)
}

func (c *client) Group() group.Group {
	return group.New(c.x, c.c.GetBucketName(), c.c.GetRegion(), c.i, c.s)
}

func (c *client) Object() object.Object {
	return object.New(c.x, c.c.GetBucketName(), c.c.GetRegion(), c.i, c.s)
}

func (c *client) Policy() policy.Policy {
	return policy.New(c.x, c.c.GetBucketName(), c.c.GetRegion(), c.i, c.s)
}

func (c *client) Role() role.Role {
	return role.New(c.x, c.c.GetBucketName(), c.c.GetRegion(), c.i, c.s)
}

func (c *client) User() user.User {
	return user.New(c.x, c.c.GetBucketName(), c.c.GetRegion(), c.i, c.s)
}

func (c *client) GetBucketName() string {
	return c.c.GetBucketName()
}

func (c *client) SetBucketName(bucket string) {
	c.c.SetBucketName(bucket)
}
