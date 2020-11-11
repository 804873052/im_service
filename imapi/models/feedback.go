package models

import "time"

type Feedback struct {
	Uid int64 `orm:"pk" json:"uid"`
	Appid int64 `json:"appid"`
	Phone string `json:"phone"`
	Content string `json:"content"`
	CreateTime time.Time `orm:"-" json:"create_time,-"`
	ModifyTime time.Time `orm:"-" json:"modify_time,-"`
}

//指定表名
func (f *Feedback) TableName() string {
	return TableName("feedback")
}
