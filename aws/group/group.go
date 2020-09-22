package group

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/nabbar/golib/errors"
)

func (cli *client) List() (map[string]string, errors.Error) {
	req := cli.iam.ListGroupsRequest(&iam.ListGroupsInput{})
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	if out, err := req.Send(cli.GetContext()); err != nil {
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
	req := cli.iam.CreateGroupRequest(&iam.CreateGroupInput{
		GroupName: aws.String(groupName),
	})
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	_, err := req.Send(cli.GetContext())

	return cli.GetError(err)
}

func (cli *client) Remove(groupName string) errors.Error {
	req := cli.iam.DeleteGroupRequest(&iam.DeleteGroupInput{
		GroupName: aws.String(groupName),
	})
	defer cli.Close(req.HTTPRequest, req.HTTPResponse)

	_, err := req.Send(cli.GetContext())

	return cli.GetError(err)
}
