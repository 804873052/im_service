package models

import (
	"github.com/astaxie/beego/orm"
)

type UserInfo struct {
	Uid        int64     `orm:"pk" json:"uid"`
	Appid      int64     `json:"appid"`
	Phone      string    `json:"phone"`
	Nickname   string    `json:"nickname"`
	HeadImg    string    `json:"headimg"`
	Sex        int8      `json:"sex"`
	Signature  string    `json:"signature"`
	RegionId   int     `json:"region_id"`
	//CreateTime time.Time `orm:"-" json:"create_time,omitempty"`
	//ModifyTime time.Time `orm:"-" json:"modify_time,omitempty"`
}

//指定表名
func (u *UserInfo) TableName() string {
	return TableName("user_info")
}

// 设置引擎为 INNODB
func (u *UserInfo) TableEngine() string {
	return "INNODB"
}

func (u *UserInfo) GetUserByUid() {
	o := orm.NewOrm()
	o.Read(u)
}

//通过手机号码获取用户信息
func (u *UserInfo) GetUserInfoByPhone() (UserInfo, error) {
	oneUser := UserInfo{}
	o := orm.NewOrm()
	qs := o.QueryTable(u)
	err := qs.Filter("Phone", u.Phone).Filter("Appid", u.Appid).One(&oneUser)
	return oneUser, err
}

//判断是否注册 已注册: 1； 未注册： 0
func (u *UserInfo) IsReg() (int64, error) {
	o := orm.NewOrm()
	qs := o.QueryTable(u)
	return qs.Filter("Phone", u.Phone).Filter("Appid", u.Appid).Count()
}

//编辑个性签名
func (u *UserInfo) EditSignatrue(signature string) (int64, error) {

	o := orm.NewOrm()
	qs := o.QueryTable(u)
	num, err := qs.Filter("Uid", u.Uid).Filter("Appid", u.Appid).Update(orm.Params{
		"signature": signature,
	})
	return num, err
}

//获取所有的用户
func GetAllUser(appid int64) ([]UserInfo, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("user_info")
	var userList []UserInfo
	_, err := qs.Filter("Appid", appid).All(&userList, "uid", "phone", "nickname", "HeadImg")
	return userList, err
}

func GetUserByUid(appid int64, uids []int64) ([]UserInfo, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("user_info")

	var userInfoList []UserInfo
	_, err := qs.Filter("appid", appid).Filter("uid__in", uids).All(&userInfoList, "uid", "nickname")
	return userInfoList, err
}

func SearchByNickname(appid int64, nickname string) ([]UserInfo, error) {

	o := orm.NewOrm()
	qs := o.QueryTable("user_info")

	var userInfoList []UserInfo
	_, err := qs.Filter("appid", appid).Filter("nickname__contains", nickname).All(&userInfoList, "uid", "nickname","headimg","sex","signature")
	return userInfoList, err

}