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
	"context"

	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	sdkstp "github.com/aws/aws-sdk-go-v2/service/s3/types"
	libhlp "github.com/nabbar/golib/aws/helper"
)

// client implements the Bucket interface for S3 bucket operations.
type client struct {
	libhlp.Helper
	iam *sdkiam.Client
	s3  *sdksss.Client
}

// WalkFunc is a callback function used by Walk to iterate over buckets.
// Return true to continue iteration, false to stop.
type WalkFunc func(bucket sdkstp.Bucket) bool

// Bucket provides operations for S3 bucket management including:
// - Lifecycle management (create, delete, list)
// - Versioning configuration
// - CORS configuration
// - Access control lists (ACL)
// - Replication configuration
// - Lifecycle rules
// - Website hosting
// - Object locking
type Bucket interface {
	// Check validates the bucket configuration and credentials.
	Check() error

	// List returns all buckets in the account.
	List() ([]sdkstp.Bucket, error)

	// Walk iterates over all buckets, calling f for each.
	// Iteration stops when f returns false or all buckets are processed.
	Walk(f WalkFunc) error

	// Create creates a new bucket with the specified region constraint.
	// If RegionConstraint is empty, uses the default region from configuration.
	Create(RegionConstraint string) error

	// CreateWithLock creates a new bucket with object locking enabled.
	// Object locking cannot be disabled after bucket creation.
	CreateWithLock(RegionConstraint string) error

	// Delete removes the bucket. Bucket must be empty before deletion.
	Delete() error

	// SetVersioning enables or disables versioning for the bucket.
	// When enabled, S3 keeps all versions of objects.
	SetVersioning(state bool) error

	// GetVersioning returns the current versioning status ("Enabled", "Suspended", or "").
	GetVersioning() (string, error)

	// LoadReplication retrieves the current replication configuration.
	LoadReplication() (*sdkstp.ReplicationConfiguration, error)

	// EnableReplication configures cross-region or same-region replication.
	// Requires source and destination role ARNs and destination bucket name.
	EnableReplication(srcRoleARN, dstRoleARN, dstBucketName string) error

	// DeleteReplication removes the replication configuration from the bucket.
	DeleteReplication() error

	// PutWebsite configures the bucket for static website hosting.
	// index: index document (e.g., "index.html")
	// error: error document (e.g., "error.html")
	PutWebsite(index, error string) error

	// GetWebsite retrieves the website configuration for the bucket.
	GetWebsite() (*sdksss.GetBucketWebsiteOutput, error)

	// SetCORS configures Cross-Origin Resource Sharing rules for the bucket.
	SetCORS(cors []sdkstp.CORSRule) error

	// GetCORS retrieves the current CORS rules for the bucket.
	GetCORS() ([]sdkstp.CORSRule, error)

	// GetACL retrieves the Access Control List (ACL) for the bucket.
	GetACL() (*sdkstp.AccessControlPolicy, error)

	// SetACL sets the ACL using either a policy, canned ACL, or headers.
	// Only one method should be provided (others can be nil/empty).
	SetACL(ACP *sdkstp.AccessControlPolicy, cannedACL sdkstp.BucketCannedACL, header ACLHeaders) error

	// SetACLPolicy sets the ACL using an Access Control Policy.
	SetACLPolicy(ACP *sdkstp.AccessControlPolicy) error

	// SetACLHeader sets the ACL using canned ACL and optional headers.
	SetACLHeader(cannedACL sdkstp.BucketCannedACL, header ACLHeaders) error

	// GetLifeCycle retrieves the lifecycle rules for the bucket.
	GetLifeCycle() ([]sdkstp.LifecycleRule, error)

	// SetLifeCycle configures lifecycle rules for automatic object management.
	// Rules define when objects transition to different storage classes or expire.
	SetLifeCycle(rules ...sdkstp.LifecycleRule) error

	// GetLock retrieves the object lock configuration for the bucket.
	GetLock() (*sdkstp.ObjectLockConfiguration, error)

	// SetLock configures object locking to prevent object deletion or modification.
	// token: optional version ID to enable lock on existing bucket
	SetLock(cfg sdkstp.ObjectLockConfiguration, token string) error
}

// New creates a new Bucket client for the specified bucket.
// The client provides access to all bucket management operations.
func New(ctx context.Context, bucket, region string, iam *sdkiam.Client, s3 *sdksss.Client) Bucket {
	return &client{
		Helper: libhlp.New(ctx, bucket, region),
		iam:    iam,
		s3:     s3,
	}
}
