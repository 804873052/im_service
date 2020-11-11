package models

import (
	"github.com/astaxie/beego/orm"
	"time"
)

type UserAuth struct {
	Id int64 `json:"id"`
	Appid int64 `json:"appid"`
	Phone string `json:"phone"`
	Password string `json:" password"`
	QqCode string `json:"qq_code,omitempty"`
	WxCode  string `json:"wx_code,omitempty"`
	CreateTime time.Time `orm:"-" json:"create_time,-"`
	ModifyTime time.Time `orm:"-" json:"modify_time,-"`
}


//指定表名
func (u *UserAuth) TableName() string {
	return TableName("user_auth")
}

// 设置引擎为 INNODB
func (u *UserAuth) TableEngine() string {
	return "INNODB"
}

func (u *UserAuth) GetUserByUid(){
	o := orm.NewOrm()
	o.Read(u)
}

func (u *UserAuth) GetUserByPhone() (int64,error) {
	o := orm.NewOrm()
	qs := o.QueryTable(u)
	return qs.Filter("Phone", u.Phone).Filter("Appid", u.Appid).Count()
}