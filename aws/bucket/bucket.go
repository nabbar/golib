package bucket

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nabbar/golib/aws/helper"
	"github.com/nabbar/golib/errors"
)

func (cli *client) Check() errors.Error {
	req := cli.s3.HeadBucketRequest(&s3.HeadBucketInput{
		Bucket: cli.GetBucketAws(),
	})

	out, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	if err != nil {
		return cli.GetError(err)
	}

	if out == nil || out.HeadBucketOutput == nil {
		//nolint #goerr113
		return helper.ErrorBucketNotFound.ErrorParent(fmt.Errorf("bucket: %s", cli.GetBucketName()))
	}

	return nil
}

func (cli *client) Create() errors.Error {
	req := cli.s3.CreateBucketRequest(&s3.CreateBucketInput{
		Bucket: cli.GetBucketAws(),
	})

	_, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	return cli.GetError(err)
}

func (cli *client) Delete() errors.Error {
	req := cli.s3.DeleteBucketRequest(&s3.DeleteBucketInput{
		Bucket: cli.GetBucketAws(),
	})

	_, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	return cli.GetError(err)
}

func (cli *client) List() ([]s3.Bucket, errors.Error) {
	req := cli.s3.ListBucketsRequest(nil)

	out, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	if err != nil {
		return make([]s3.Bucket, 0), cli.GetError(err)
	}

	if out == nil || out.Buckets == nil {
		return make([]s3.Bucket, 0), helper.ErrorAwsEmpty.Error(nil)
	}

	return out.Buckets, nil
}

func (cli *client) SetVersioning(state bool) errors.Error {
	var status s3.BucketVersioningStatus = helper.STATE_ENABLED
	if !state {
		status = helper.STATE_SUSPENDED
	}

	vConf := s3.VersioningConfiguration{
		Status: status,
	}
	input := s3.PutBucketVersioningInput{
		Bucket:                  cli.GetBucketAws(),
		VersioningConfiguration: &vConf,
	}

	req := cli.s3.PutBucketVersioningRequest(&input)
	_, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	return cli.GetError(err)
}

func (cli *client) GetVersioning() (string, errors.Error) {
	input := s3.GetBucketVersioningInput{
		Bucket: cli.GetBucketAws(),
	}

	req := cli.s3.GetBucketVersioningRequest(&input)
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	out, err := req.Send(cli.GetContext())

	if err != nil {
		return "", cli.GetError(err)
	}

	// MarshalValue always return error as nil
	v, _ := out.Status.MarshalValue()

	return v, nil
}

func (cli *client) EnableReplication(srcRoleARN, dstRoleARN, dstBucketName string) errors.Error {
	var status s3.ReplicationRuleStatus = helper.STATE_ENABLED

	replicationConf := s3.ReplicationConfiguration{
		Role: aws.String(srcRoleARN + "," + dstRoleARN),
		Rules: []s3.ReplicationRule{
			{
				Destination: &s3.Destination{
					Bucket: aws.String("arn:aws:s3:::" + dstBucketName),
				},
				Status: status,
				Prefix: aws.String(""),
			},
		},
	}

	req := cli.s3.PutBucketReplicationRequest(&s3.PutBucketReplicationInput{
		Bucket:                   cli.GetBucketAws(),
		ReplicationConfiguration: &replicationConf,
	})
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	_, err := req.Send(cli.GetContext())

	return cli.GetError(err)
}

func (cli *client) DeleteReplication() errors.Error {
	req := cli.s3.DeleteBucketReplicationRequest(&s3.DeleteBucketReplicationInput{
		Bucket: cli.GetBucketAws(),
	})
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	_, err := req.Send(cli.GetContext())

	return cli.GetError(err)
}
