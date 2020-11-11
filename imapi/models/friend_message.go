package models

import (
	"github.com/astaxie/beego/orm"
	"time"
)

type FriendMessage struct {
	RequestId             int64     `orm:"pk" json:"request_id"`
	Uid            int64     `json:"uid"`
	MsgContent      string     `json:"msg_content"`
	CreateTime     time.Time `orm:"-" json:"create_time,-"`
	ModifyTime     time.Time `orm:"-" json:"modify_time,-"`
}

//指定表名
func (fr *FriendMessage) TableName() string {
	return TableName("request_message")
}

// 设置引擎为 INNODB
func (fr *FriendMessage) TableEngine() string {
	return "INNODB"
}


func (fr *FriendMessage) GetUserByUid() {
	o := orm.NewOrm()
	o.Read(fr)
}

/**
 添加好友申请消息
 */
func  InsertRecordMessage(requestId int64, uid int64, msg string, o orm.Ormer)  (int64, error) {
	var frqData FriendMessage
	frqData.RequestId = requestId
	frqData.Uid = uid
	frqData.MsgContent = msg

	id, err := o.Insert(&frqData)

	return id, err
}
