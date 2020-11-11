package models

import (
	"errors"
	"github.com/astaxie/beego/orm"
)

type MyPage struct {
	Uid int64 `json:"uid"`
	Nickname string `json:"nickname"`
	HeadImg string `json:"headimg"`
	Signature string `json:"signature"`
	RegionId int `json:"region_id"`
	RegionDesc string `json:"region_desc"`
}

func UpdateMyInfo(uid int64,nickname string,sex int8,headimg string,regionId int,signature string) (int64,error) {
	o := orm.NewOrm()
	qs := o.QueryTable("user_info")
	updateParams := orm.Params{}
	if nickname != "" {
		updateParams["nickname"] = nickname
	}
	if sex != 0 {
		updateParams["sex"] = sex
	}
	if headimg != "" {
		updateParams["headimg"] = headimg
	}
	if regionId != 0 {
		updateParams["regionId"] = regionId
	}
	if signature != "" {
		updateParams["signature"] = signature
	}
	if len(updateParams) > 0 {
		return qs.Filter("Uid", uid).Update(updateParams)
	}else {
		return 0,errors.New("未提供任何参数！")
	}


}