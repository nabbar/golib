package policy

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/nabbar/golib/errors"
)

func (cli *client) List() (map[string]string, errors.Error) {
	out, err := cli.iam.ListPolicies(cli.GetContext(), &iam.ListPoliciesInput{})

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
	out, err := cli.iam.CreatePolicy(cli.GetContext(), &iam.CreatePolicyInput{
		PolicyName:     aws.String(name),
		Description:    aws.String(desc),
		PolicyDocument: aws.String(policy),
	})

	if err != nil {
		return "", cli.GetError(err)
	} else {
		return *out.Policy.Arn, nil
	}
}

func (cli *client) Update(polArn, polContents string) errors.Error {
	out, err := cli.iam.ListPolicyVersions(cli.GetContext(), &iam.ListPolicyVersionsInput{
		PolicyArn: aws.String(polArn),
	})

	if err != nil {
		return cli.GetError(err)
	} else {
		for _, v := range out.Versions {
			if cli.GetContext().Err() != nil {
				return nil
			}

			if !*v.IsDefaultVersion {
				_, _ = cli.iam.DeletePolicyVersion(cli.GetContext(), &iam.DeletePolicyVersionInput{
					PolicyArn: aws.String(polArn),
					VersionId: v.VersionId,
				})
			}
		}
	}

	if cli.GetContext().Err() != nil {
		return nil
	}

	_, err = cli.iam.CreatePolicyVersion(cli.GetContext(), &iam.CreatePolicyVersionInput{
		PolicyArn:      aws.String(polArn),
		PolicyDocument: aws.String(polContents),
		SetAsDefault:   aws.Bool(true),
	})

	return cli.GetError(err)
}

func (cli *client) Delete(polArn string) errors.Error {
	out, err := cli.iam.ListPolicyVersions(cli.GetContext(), &iam.ListPolicyVersionsInput{
		PolicyArn: aws.String(polArn),
	})

	if err != nil {
		return cli.GetError(err)
	} else {
		for _, v := range out.Versions {
			if cli.GetContext().Err() != nil {
				return nil
			}

			if !*v.IsDefaultVersion {
				_, _ = cli.iam.DeletePolicyVersion(cli.GetContext(), &iam.DeletePolicyVersionInput{
					PolicyArn: aws.String(polArn),
					VersionId: v.VersionId,
				})
			}
		}
	}

	if cli.GetContext().Err() != nil {
		return nil
	}

	_, err = cli.iam.DeletePolicy(cli.GetContext(), &iam.DeletePolicyInput{
		PolicyArn: aws.String(polArn),
	})

	return cli.GetError(err)
}
