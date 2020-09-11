package policy

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/nabbar/golib/errors"
)

func (cli *client) List() (map[string]string, errors.Error) {
	req := cli.iam.ListPoliciesRequest(&iam.ListPoliciesInput{})

	out, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	if err != nil {
		return nil, cli.GetError(err)
	} else {
		var res = make(map[string]string)

		for _, p := range out.Policies {
			res[*p.PolicyName] = *p.Arn
		}

		return res, nil
	}
}

func (cli *client) Add(name, desc, policy string) (string, errors.Error) {
	req := cli.iam.CreatePolicyRequest(&iam.CreatePolicyInput{
		PolicyName:     aws.String(name),
		Description:    aws.String(desc),
		PolicyDocument: aws.String(policy),
	})

	out, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	if err != nil {
		return "", cli.GetError(err)
	} else {
		return *out.Policy.Arn, nil
	}
}

func (cli *client) Update(polArn, polContents string) errors.Error {
	req := cli.iam.ListPolicyVersionsRequest(&iam.ListPolicyVersionsInput{
		PolicyArn: aws.String(polArn),
	})

	out, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	if err != nil {
		return cli.GetError(err)
	} else {
		for _, v := range out.Versions {
			if cli.GetContext().Err() != nil {
				return nil
			}

			if !*v.IsDefaultVersion {
				reqD := cli.iam.DeletePolicyVersionRequest(&iam.DeletePolicyVersionInput{
					PolicyArn: aws.String(polArn),
					VersionId: v.VersionId,
				})

				if o, e := reqD.Send(cli.GetContext()); e != nil {
					continue
				} else if o == nil {
					continue
				}
			}
		}
	}

	reqG := cli.iam.CreatePolicyVersionRequest(&iam.CreatePolicyVersionInput{
		PolicyArn:      aws.String(polArn),
		PolicyDocument: aws.String(polContents),
		SetAsDefault:   aws.Bool(true),
	})

	if cli.GetContext().Err() != nil {
		return nil
	}

	_, err = reqG.Send(cli.GetContext())
	defer cli.Close(reqG.HTTPRequest, reqG.HTTPResponse)

	return cli.GetError(err)
}

func (cli *client) Delete(polArn string) errors.Error {
	req := cli.iam.ListPolicyVersionsRequest(&iam.ListPolicyVersionsInput{
		PolicyArn: aws.String(polArn),
	})

	out, err := req.Send(cli.GetContext())
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	if err != nil {
		return cli.GetError(err)
	} else {
		for _, v := range out.Versions {
			if cli.GetContext().Err() != nil {
				return nil
			}

			if !*v.IsDefaultVersion {
				reqD := cli.iam.DeletePolicyVersionRequest(&iam.DeletePolicyVersionInput{
					PolicyArn: aws.String(polArn),
					VersionId: v.VersionId,
				})

				if o, e := reqD.Send(cli.GetContext()); e != nil {
					continue
				} else if o == nil {
					continue
				}
			}
		}
	}

	if cli.GetContext().Err() != nil {
		return nil
	}

	reqG := cli.iam.DeletePolicyRequest(&iam.DeletePolicyInput{
		PolicyArn: aws.String(polArn),
	})

	_, err = reqG.Send(cli.GetContext())
	defer cli.Close(reqG.HTTPRequest, reqG.HTTPResponse)

	return cli.GetError(err)
}
