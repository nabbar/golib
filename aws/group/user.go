package group

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/nabbar/golib/errors"
)

func (cli *client) UserCheck(username, groupName string) (errors.Error, bool) {
	req := cli.iam.ListGroupsForUserRequest(&iam.ListGroupsForUserInput{
		UserName: aws.String(username),
	})
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	if out, err := req.Send(cli.GetContext()); err != nil {
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
	req := cli.iam.ListGroupsForUserRequest(&iam.ListGroupsForUserInput{
		UserName: aws.String(username),
	})
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	if out, err := req.Send(cli.GetContext()); err != nil {
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
	req := cli.iam.AddUserToGroupRequest(&iam.AddUserToGroupInput{
		UserName:  aws.String(username),
		GroupName: aws.String(groupName),
	})
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	_, err := req.Send(cli.GetContext())

	return cli.GetError(err)
}

func (cli *client) UserRemove(username, groupName string) errors.Error {
	req := cli.iam.RemoveUserFromGroupRequest(&iam.RemoveUserFromGroupInput{
		UserName:  aws.String(username),
		GroupName: aws.String(groupName),
	})
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	_, err := req.Send(cli.GetContext())

	return cli.GetError(err)
}
