package group

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/nabbar/golib/errors"
)

func (cli *client) UserCheck(username, groupName string) (errors.Error, bool) {
	out, err := cli.iam.ListGroupsForUser(cli.GetContext(), &iam.ListGroupsForUserInput{
		UserName: aws.String(username),
	})

	if err != nil {
		return cli.GetError(err), false
	} else {
		for _, g := range out.Groups {
			if *g.GroupName == groupName {
				return nil, true
			}
		}
	}

	return nil, false
}

func (cli *client) UserList(username string) ([]string, errors.Error) {
	out, err := cli.iam.ListGroupsForUser(cli.GetContext(), &iam.ListGroupsForUserInput{
		UserName: aws.String(username),
	})

	if err != nil {
		return nil, cli.GetError(err)
	} else {
		var res = make([]string, 0)

		for _, g := range out.Groups {
			res = append(res, *g.GroupName)
		}

		return res, nil
	}
}

func (cli *client) UserAdd(username, groupName string) errors.Error {
	_, err := cli.iam.AddUserToGroup(cli.GetContext(), &iam.AddUserToGroupInput{
		UserName:  aws.String(username),
		GroupName: aws.String(groupName),
	})

	return cli.GetError(err)
}

func (cli *client) UserRemove(username, groupName string) errors.Error {
	_, err := cli.iam.RemoveUserFromGroup(cli.GetContext(), &iam.RemoveUserFromGroupInput{
		UserName:  aws.String(username),
		GroupName: aws.String(groupName),
	})

	return cli.GetError(err)
}
