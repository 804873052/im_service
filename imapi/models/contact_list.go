package models

import (
	"errors"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"strings"
)

/**
  我的通讯录列表
*/
func GetMyContactList(uid int64) ([]UserInfo,error) {

	o := orm.NewOrm()
	qs := o.QueryTable("friend")
	friendUids := []Friend{}
	num,_ := qs.Filter("uid",uid).Distinct().All(&friendUids,"friend_uid")
	if num == 0 {
		return nil,errors.New("暂无通讯录信息！")
	}
	logs.Debug(friendUids)
	//加入的组的id切片数组
	fUids := make([]int64,0)
	for _,v := range friendUids {
		fUids = append(fUids,v.FriendUid)
	}
	logs.Debug(fUids)

	contactList := []UserInfo{}
	qs2 := o.QueryTable("user_info")
	qs2.Filter("uid__in", fUids).All(&contactList)
	return contactList,nil
}

/**
  是否为我的好友
*/
func IsMyFriend(uid int64,friendUid int64) (int64,error) {

	o := orm.NewOrm()
	qs := o.QueryTable("friend")
	return qs.Filter("uid",uid).Filter("friend_uid",friendUid).Count()

}

/**
  获取地区的名称
*/
func GetRegoinName(regionId int) (string,error) {
	region := Region{}
	o := orm.NewOrm()
	qs := o.QueryTable("region")
	err := qs.Filter("id",regionId).One(&region,"path")
	if err != nil {
		return "", err
	}
	trimPath := strings.Trim(region.Path,",")
	//logs.Info(trimPath)
	pathSlice := strings.Split(trimPath, ",")
	//logs.Info(pathSlice)
	regionSlice := []Region{}
	qs.Filter("id__in",pathSlice).All(&regionSlice,"name")
	logs.Info(regionSlice)
	var regionString string
	for _,v := range regionSlice {
		if (v.Name != "亚洲") {
			regionString += v.Name + " "
		}
	}
	regionString = strings.Trim(regionString," ")
	return regionString, nil
}