package object

import (
	"regexp"

	"github.com/nabbar/golib/errors"
)

func (cli *client) Find(pattern string) ([]string, errors.Error) {
	var (
		result = make([]string, 0)
		token  = ""
	)

	for {
		if lst, tok, cnt, err := cli.List(token); err != nil {
			return result, cli.GetError(err)
		} else if cnt > 0 {
			token = tok
			for _, o := range lst {
				if ok, _ := regexp.MatchString(pattern, *o.Key); ok {
					result = append(result, *o.Key)
				}
			}
		} else {
			return result, nil
		}

		if token == "" {
			return result, nil
		}
	}
}
