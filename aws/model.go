package aws

import (
	"github.com/nabbar/golib/aws/bucket"
	"github.com/nabbar/golib/aws/group"
	"github.com/nabbar/golib/aws/object"
	"github.com/nabbar/golib/aws/policy"
	"github.com/nabbar/golib/aws/role"
	"github.com/nabbar/golib/aws/user"
	"github.com/nabbar/golib/errors"
)

func (c *client) ForcePathStyle(enabled bool) errors.Error {
	c.p = enabled

	if s, e := c.newClientS3(nil); e != nil {
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
	return bucket.New(c.x, c.c.GetBucketName(), c.i, c.s)
}

func (c *client) Group() group.Group {
	return group.New(c.x, c.c.GetBucketName(), c.i, c.s)
}

func (c *client) Object() object.Object {
	return object.New(c.x, c.c.GetBucketName(), c.i, c.s)
}

func (c *client) Policy() policy.Policy {
	return policy.New(c.x, c.c.GetBucketName(), c.i, c.s)
}

func (c *client) Role() role.Role {
	return role.New(c.x, c.c.GetBucketName(), c.i, c.s)
}

func (c *client) User() user.User {
	return user.New(c.x, c.c.GetBucketName(), c.i, c.s)
}

func (c *client) GetBucketName() string {
	return c.c.GetBucketName()
}

func (c *client) SetBucketName(bucket string) {
	c.c.SetBucketName(bucket)
}
