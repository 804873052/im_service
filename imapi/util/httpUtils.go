package util

import (
	"encoding/json"
	"github.com/astaxie/beego/httplib"
)

func Post(data interface{}, url string) (*httplib.BeegoHTTPRequest, error) {
	req := httplib.Post(url)
	jsons, err := json.Marshal(data)
	if nil != err {
		return nil, err
	}
	req.Body(jsons)
	return req, nil
}
