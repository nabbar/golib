/*
 *  MIT License
 *
 *  Copyright (c) 2021 Nicolas JUHEL
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

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
	sdksss "github.com/aws/aws-sdk-go-v2/service/s3"
	awsedp "github.com/aws/smithy-go/endpoints"
)

type resolverIam struct {
	r func(service, region string) (sdkaws.Endpoint, error) // nolint
}

func (r *resolverIam) ResolveEndpoint(region string, options sdkiam.EndpointResolverOptions) (sdkaws.Endpoint, error) { // nolint
	return r.r("iam", region)
}

type resolverIamV2 struct {
	r func(service, region string) (awsedp.Endpoint, error)
}

func (r *resolverIamV2) ResolveEndpoint(ctx context.Context, params sdkiam.EndpointParameters) (awsedp.Endpoint, error) {
	return r.r("iam", *params.Region)
}

type resolverS3 struct {
	r func(service, region string) (sdkaws.Endpoint, error) // nolint
}

func (r *resolverS3) ResolveEndpoint(region string, options sdksss.EndpointResolverOptions) (sdkaws.Endpoint, error) { // nolint
	return r.r("s3", region)
}

type resolverS3V2 struct {
	r func(service, region string) (awsedp.Endpoint, error)
}

func (r *resolverS3V2) ResolveEndpoint(ctx context.Context, params sdksss.EndpointParameters) (awsedp.Endpoint, error) {
	return r.r("s3", *params.Region)
}

func (c *client) _NewIAMResolver(cfg *sdkaws.Config) sdkiam.EndpointResolver {
	return &resolverIam{
		r: cfg.EndpointResolver.ResolveEndpoint, // nolint
	}
}

func (c *client) _NewIAMResolverV2(cfg Config) sdkiam.EndpointResolverV2 {
	return &resolverIamV2{
		r: func(service, region string) (awsedp.Endpoint, error) {
			edp, err := cfg.ResolveEndpoint(service, region)
			if err != nil {
				return awsedp.Endpoint{}, err
			}

			uri, err := url.Parse(edp.URL)
			if err != nil {
				return awsedp.Endpoint{}, err
			}

			return awsedp.Endpoint{
				URI: *uri,
			}, nil
		},
	}
}

func (c *client) _NewS3Resolver(cfg *sdkaws.Config) sdksss.EndpointResolver {
	return &resolverS3{
		r: cfg.EndpointResolver.ResolveEndpoint, // nolint
	}
}

func (c *client) _NewS3ResolverV2(cfg Config) sdksss.EndpointResolverV2 {
	return &resolverS3V2{
		r: func(service, region string) (awsedp.Endpoint, error) {
			edp, err := cfg.ResolveEndpoint(service, region)
			if err != nil {
				return awsedp.Endpoint{}, err
			}

			uri, err := url.Parse(edp.URL)
			if err != nil {
				return awsedp.Endpoint{}, err
			}

			return awsedp.Endpoint{
				URI: *uri,
			}, nil
		},
	}
}
