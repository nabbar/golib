package group

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/nabbar/golib/errors"
)

func (cli *client) List() (map[string]string, errors.Error) {
	if out, err := cli.iam.ListGroups(cli.GetContext(), &iam.ListGroupsInput{}); err != nil {
		return nil, cli.GetError(err)
	} else {
		var res = make(map[string]string)

		for _, g := range out.Groups {
			res[*g.GroupId] = *g.GroupName
		}

		return res, nil
	}
}

func (cli *client) Add(groupName string) errors.Error {
	_, err := cli.iam.CreateGroup(cli.GetContext(), &iam.CreateGroupInput{
		GroupName: aws.String(groupName),
	})

	return cli.GetError(err)
}

func (cli *client) Remove(groupName string) errors.Error {
	_, err := cli.iam.DeleteGroup(cli.GetContext(), &iam.DeleteGroupInput{
		GroupName: aws.String(groupName),
	})

	return cli.GetError(err)
}
