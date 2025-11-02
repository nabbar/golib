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
	"net/url"
	"sync"
	"time"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdksv4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	awsbck "github.com/nabbar/golib/aws/bucket"
	awsgrp "github.com/nabbar/golib/aws/group"
	awshlp "github.com/nabbar/golib/aws/helper"
	awsobj "github.com/nabbar/golib/aws/object"
	awspol "github.com/nabbar/golib/aws/policy"
	awsrol "github.com/nabbar/golib/aws/role"
	awsusr "github.com/nabbar/golib/aws/user"
	libhtc "github.com/nabbar/golib/httpcli"
)

type Config interface {
	// Check checks the configuration and returns an error if it's not valid.
	//
	// The function is safe for concurrent use by multiple goroutines.
	//
	// The context must be non-nil. If the context is canceled or timed out,
	// the function will return ctx.Err() immediately.
	Check(ctx context.Context) error
	// Validate validates the configuration and returns an error if it's not valid.
	//
	// The function is safe for concurrent use by multiple goroutines.
	//
	// The function will return an error if the configuration is not valid.
	// The error will be of type aws.ConfigValidationError.
	//
	// The function will return an error if the region is not set.
	// The function will return an error if the access key is not set.
	// The function will return an error if the secret key is not set.
	//
	// The function is idempotent.
	Validate() error

	// GetAccessKey returns the access key for the configuration.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function will return an empty string if the access key is not set.
	//
	// The function is idempotent.
	GetAccessKey() string
	// GetSecretKey returns the secret key for the configuration.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function will return an empty string if the secret key is not set.
	// The function is idempotent.
	GetSecretKey() string
	// SetCredentials sets the access key and secret key for the configuration.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function will return an error if the access key or secret key are empty.
	// The function is idempotent.
	//
	// Parameters:
	// accessKey - the access key to set for the configuration.
	// secretKey - the secret key to set for the configuration.
	//
	// Returns:
	// error - an error if the access key or secret key are empty.
	SetCredentials(accessKey, secretKey string)
	// ResetRegionEndpoint resets the region endpoint for the configuration.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function is idempotent.
	// The function does not return any error.
	// The function does not block.
	// The function does not allocate any memory.
	// The function does not depend on any external state.
	ResetRegionEndpoint()
	// RegisterRegionEndpoint registers a region endpoint for the configuration.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function will return an error if the region is empty.
	// The function will return an error if the endpoint is empty.
	// The function is idempotent.
	//
	// Parameters:
	// region - the region to register for the configuration.
	// endpoint - the endpoint to register for the configuration.
	//
	// Returns:
	// error - an error if the region or endpoint are empty.
	RegisterRegionEndpoint(region string, endpoint *url.URL) error
	// RegisterRegionAws registers a region endpoint for the configuration.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function will return an error if the region is empty.
	// The function will return an error if the endpoint is empty.
	// The function is idempotent.
	//
	// Parameters:
	// endpoint - the endpoint to register for the configuration.
	//
	// Returns:
	// error - an error if the region or endpoint are empty.
	RegisterRegionAws(endpoint *url.URL) error
	// SetRegion sets the region for the configuration.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function is idempotent.
	//
	// Parameters:
	// region - the region to set for the configuration.
	//
	// Returns:
	// None
	SetRegion(region string)
	// GetRegion returns the region for the configuration.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function is idempotent.
	//
	// Returns:
	// string - the region for the configuration.
	GetRegion() string
	// SetEndpoint sets the endpoint for the configuration.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function is idempotent.
	//
	// Parameters:
	// endpoint - the endpoint to set for the configuration.
	//
	// Returns:
	// None
	SetEndpoint(endpoint *url.URL)
	// GetEndpoint returns the endpoint for the configuration.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function is idempotent.
	//
	// Returns:
	// *url.URL - the endpoint for the configuration.
	//
	GetEndpoint() *url.URL

	// IsHTTPs returns true if the configuration is set to use HTTPS.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function is idempotent.
	//
	// Returns:
	// bool - true if the configuration is set to use HTTPS, false otherwise.
	IsHTTPs() bool
	// ResolveEndpoint resolves an endpoint for a given service and region.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function will return an error if the service or region are empty.
	//
	// Parameters:
	// service - the service to resolve the endpoint for.
	// region - the region to resolve the endpoint for.
	//
	// Returns:
	// sdkaws.Endpoint - the resolved endpoint.
	// error - an error if the service or region are empty.
	ResolveEndpoint(service, region string) (sdkaws.Endpoint, error) // nolint
	// ResolveEndpointWithOptions resolves an endpoint for a given service and region
	// with the provided options.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function will return an error if the service or region are empty.
	//
	// Parameters:
	// service - the service to resolve the endpoint for.
	// region - the region to resolve the endpoint for.
	// options - zero or more options to use when resolving the endpoint.
	//
	// Returns:
	// sdkaws.Endpoint - the resolved endpoint.
	// error - an error if the service or region are empty.
	ResolveEndpointWithOptions(service, region string, options ...interface{}) (sdkaws.Endpoint, error) // nolint
	// GetDisableHTTPS returns true if the configuration is set to disable
	// HTTPS. The function is safe for concurrent use by multiple
	// goroutines. The function is idempotent.
	//
	// Returns:
	// bool - true if the configuration is set to disable HTTPS,
	// false otherwise.
	GetDisableHTTPS() bool
	// GetResolvedRegion returns the resolved region for the configuration.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function is idempotent.
	//
	// Returns:
	// string - the resolved region for the configuration.
	GetResolvedRegion() string
	// SetRetryer sets the retryer for the configuration.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function is idempotent.
	//
	// Parameters:
	// retryer - the retryer to set for the configuration.
	SetRetryer(retryer func() sdkaws.Retryer)

	// GetConfig returns the configuration for the AWS client.
	//
	// The function is safe for concurrent use by multiple goroutines.
	//
	// Parameters:
	// ctx - the context to use when loading the configuration.
	// cli - the HTTP client to use when loading the configuration.
	//
	// Returns:
	// *sdkaws.Config - the configuration for the AWS client.
	// error - an error if the configuration could not be loaded.
	GetConfig(ctx context.Context, cli libhtc.HttpClient) (*sdkaws.Config, error)
	// JSON returns the JSON representation of the configuration.
	//
	// The function is safe for concurrent use by multiple goroutines.
	//
	// Returns:
	// []byte - the JSON representation of the configuration.
	// error - an error if the configuration could not be JSON encoded.
	JSON() ([]byte, error)
	// Clone returns a deep copy of the configuration.
	//
	// The function is safe for concurrent use by multiple goroutines.
	//
	// Returns:
	// Config - a deep copy of the configuration.
	Clone() Config

	// GetBucketName returns the name of the bucket associated with the AWS client.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function is idempotent.
	//
	// Returns:
	// string - the name of the bucket associated with the AWS client.
	GetBucketName() string
	// SetBucketName sets the name of the bucket associated with the AWS client.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function is idempotent.
	//
	// Parameters:
	// bucket - the name of the bucket to set for the AWS client.
	SetBucketName(bucket string)

	// SetChecksumValidation sets the checksum validation for the configuration.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function is idempotent.
	//
	// Parameters:
	// req - the request checksum validation to set for the configuration.
	// rsp - the response checksum validation to set for the configuration.
	SetChecksumValidation(req sdkaws.RequestChecksumCalculation, rsp sdkaws.ResponseChecksumValidation)
	// GetChecksumValidation returns the checksum validation settings for the configuration.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function is idempotent.
	//
	// Returns:
	// req - the request checksum validation for the configuration.
	// rsp - the response checksum validation for the configuration.
	GetChecksumValidation() (req sdkaws.RequestChecksumCalculation, rsp sdkaws.ResponseChecksumValidation)
}

type AWS interface {
	// Bucket returns an AWS Bucket resource.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function is idempotent.
	//
	// Returns:
	// awsbck.Bucket - an AWS Bucket resource.
	Bucket() awsbck.Bucket
	// Group returns an AWS Group resource.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function is idempotent.
	//
	// Returns:
	// awsgrp.Group - an AWS Group resource.
	Group() awsgrp.Group
	// Object returns an AWS Object resource.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function is idempotent.
	//
	// Returns:
	// awsobj.Object - an AWS Object resource.
	Object() awsobj.Object
	// Policy returns an AWS Policy resource.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function is idempotent.
	//
	// Returns:
	// awspol.Policy - an AWS Policy resource.
	Policy() awspol.Policy
	// Role returns an AWS Role resource.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function is idempotent.
	//
	// Returns:
	// awsrol.Role - an AWS Role resource.
	Role() awsrol.Role
	// User returns an AWS User resource.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function is idempotent.
	//
	// Returns:
	// awsusr.User - an AWS User resource.
	User() awsusr.User

	// Config returns the configuration associated with the AWS client.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function is idempotent.
	//
	// Returns:
	// Config - the configuration associated with the AWS client.
	Config() Config
	// HTTPCli returns the HTTP client associated with the AWS client.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function is idempotent.
	//
	// Returns:
	// libhtc.HttpClient - the HTTP client associated with the AWS client.
	HTTPCli() libhtc.HttpClient
	// Clone returns a deep copy of the AWS client.
	//
	// The function is safe for concurrent use by multiple goroutines.
	//
	// Parameters:
	// ctx - the context to use when cloning the AWS client. If nil, context.Background() is used.
	//
	// Returns:
	// AWS - a deep copy of the AWS client.
	// error - an error if the AWS client could not be cloned.
	Clone(ctx context.Context) (AWS, error)
	// NewForConfig returns a new AWS client with the given configuration and HTTP client.
	//
	// The function is safe for concurrent use by multiple goroutines.
	//
	// Parameters:
	// ctx - the context to use when creating the AWS client. If nil, context.Background() is used.
	// cfg - the configuration for the AWS client. If nil, awshlp.ErrorConfigEmpty is returned.
	// httpClient - the HTTP client to use when creating the AWS client. If nil, the default HTTP client is used.
	//
	// Returns:
	// AWS - the new AWS client.
	// error - an error if the AWS client could not be created.
	NewForConfig(ctx context.Context, cfg Config) (AWS, error)
	// ForcePathStyle forces the path style for the configuration.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function is idempotent.
	//
	// Parameters:
	// ctx - the context to use when forcing the path style.
	// enabled - true to force path style, false otherwise.
	//
	// Returns:
	// error - an error if the path style could not be forced.
	ForcePathStyle(ctx context.Context, enabled bool) error
	// ForceSignerOptions forces the signer options for the configuration.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function is idempotent.
	//
	// Parameters:
	// ctx - the context to use when forcing the signer options.
	// fct - zero or more functions to use when forcing the signer options.
	//
	// Returns:
	// error - an error if the signer options could not be forced.
	ForceSignerOptions(ctx context.Context, fct ...func(signer *sdksv4.SignerOptions)) error

	// GetBucketName returns the name of the bucket associated with the AWS client.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function is idempotent.
	//
	// Returns:
	// string - the name of the bucket associated with the AWS client.
	GetBucketName() string
	// SetBucketName sets the name of the bucket associated with the AWS client.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function is idempotent.
	//
	// Parameters:
	// bucket - the name of the bucket to set for the AWS client.
	SetBucketName(bucket string)
	// SetHTTPTimeout sets the HTTP timeout for the configuration.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function is idempotent.
	//
	// Parameters:
	// dur - the HTTP timeout to set for the configuration.
	//
	// Returns:
	// error - an error if the HTTP timeout could not be set.
	SetHTTPTimeout(dur time.Duration) error
	// GetHTTPTimeout returns the HTTP timeout for the configuration.
	//
	// The function is safe for concurrent use by multiple goroutines.
	// The function is idempotent.
	//
	// Returns:
	// time.Duration - the HTTP timeout for the configuration.
	GetHTTPTimeout() time.Duration
	// GetClientS3 returns the S3 client for the configuration.
	//
	// The function is safe for concurrent use by multiple goroutines.
	//
	// Returns:
	// *sdksss.Client - the S3 client for the configuration.
	GetClientS3() *sdksss.Client
	// SetClientS3 sets the S3 client for the configuration.
	//
	// The function is safe for concurrent use by multiple goroutines.
	//
	// Parameters:
	// aws - the S3 client to set for the configuration.
	SetClientS3(aws *sdksss.Client)
	// GetClientIam returns the IAM client for the configuration.
	//
	// The function is safe for concurrent use by multiple goroutines.
	//
	// Returns:
	// *sdkiam.Client - the IAM client for the configuration.
	GetClientIam() *sdkiam.Client
	// SetClientIam sets the IAM client for the configuration.
	//
	// The function is safe for concurrent use by multiple goroutines.
	//
	// Parameters:
	// aws - the IAM client to set for the configuration.
	//
	// Returns:
	// error - an error if the IAM client could not be set.
	SetClientIam(aws *sdkiam.Client)
}

// New returns a new AWS client with the given configuration and HTTP client.
//
// The function is safe for concurrent use by multiple goroutines.
//
// Parameters:
// ctx - the context to use when creating the AWS client. If nil, context.Background() is used.
// cfg - the configuration for the AWS client. If nil, awshlp.ErrorConfigEmpty is returned.
// httpClient - the HTTP client to use when creating the AWS client. If nil, the default HTTP client is used.
//
// Returns:
// AWS - the new AWS client.
// error - an error if the AWS client could not be created.
func New(ctx context.Context, cfg Config, httpClient libhtc.HttpClient) (AWS, error) {
	if cfg == nil {
		return nil, awshlp.ErrorConfigEmpty.Error(nil)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	cli := &client{
		m: sync.Mutex{},
		p: false,
		o: make([]func(signer *sdksv4.SignerOptions), 0),
		x: ctx,
		c: cfg,
		i: nil,
		s: nil,
		h: httpClient,
	}

	if i, e := cli._NewClientIAM(ctx, httpClient, nil); e != nil {
		return nil, e
	} else {
		cli.i = i
	}

	if s, e := cli._NewClientS3(ctx, httpClient, nil); e != nil {
		return nil, e
	} else {
		cli.s = s
	}

	return cli, nil
}
