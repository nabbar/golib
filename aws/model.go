package aws

import (
	"github.com/nabbar/golib/aws/bucket"
	"github.com/nabbar/golib/aws/group"
	"github.com/nabbar/golib/aws/object"
	"github.com/nabbar/golib/aws/policy"
	"github.com/nabbar/golib/aws/role"
	"github.com/nabbar/golib/aws/user"
)

func (c *client) ForcePathStyle(enabled bool) {
	c.p = enabled
	c.s.ForcePathStyle = enabled
}

func (c *client) Config() Config {
	return c.c
}

func (c *client) Bucket() bucket.Bucket {
	return bucket.New(c.x, c.c.GetBucketName(), c.getCliIAM(), c.getCliS3())
}

func (c *client) Group() group.Group {
	return group.New(c.x, c.c.GetBucketName(), c.getCliIAM(), c.getCliS3())
}

func (c *client) Object() object.Object {
	return object.New(c.x, c.c.GetBucketName(), c.getCliIAM(), c.getCliS3())
}

func (c *client) Policy() policy.Policy {
	return policy.New(c.x, c.c.GetBucketName(), c.getCliIAM(), c.getCliS3())
}

func (c *client) Role() role.Role {
	return role.New(c.x, c.c.GetBucketName(), c.getCliIAM(), c.getCliS3())
}

func (c *client) User() user.User {
	return user.New(c.x, c.c.GetBucketName(), c.getCliIAM(), c.getCliS3())
}

func (c *client) GetBucketName() string {
	return c.c.GetBucketName()
}

func (c *client) SetBucketName(bucket string) {
	c.c.SetBucketName(bucket)
}
