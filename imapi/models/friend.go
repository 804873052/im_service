package models

import (
	"github.com/astaxie/beego/orm"
	"time"
)

type Friend struct {
	Appid          int64   `orm:"pk" json:"appid"`
	Uid            int64   `json:"uid"`
	FriendUid      int64   `json:"friend_uid"`
	FriendNickname *string `json:"friend_nickname"`
	Timestamp      time.Time `orm:"-" json:"timestamp,-"`
}

//指定表名
func (f *Friend) TableName() string {
	return TableName("friend")
}

// 设置引擎为 INNODB
func (f *Friend) TableEngine() string {
	return "INNODB"
}

// 多字段唯一键
func (f *Friend) TableUnique() [][]string {
	return [][]string{
		[]string{"AppId", "uid", "friend_uid"},
	}
}

func (f *Friend) GetUserByUid() {
	o := orm.NewOrm()
	o.Read(f)
}

/**
  我的好友列表
 */
func  GetMyFriendsList(uid int64) []Friend {
	var friends []Friend
	o := orm.NewOrm()
	_, err := o.Raw("SELECT * FROM friend WHERE uid = ?", uid).QueryRows(&friends)
	if err != nil {
	}
	return friends
}


/**
添加好友--添加两条记录
*/
func  InsertRecordFriend(uid int64, fuid int64, appid int64, o orm.Ormer)  error {

	qs := o.QueryTable("friend")
	num,err := qs.Filter("uid", fuid).Filter("friend_uid", uid).Count()

	//之前添加过好友 —— 存在对方是我的好友记录
	if(num != 0) {
		_, err = o.Insert(&Friend{Appid: appid, Uid: uid, FriendUid: fuid})
	//之前未添加过好友 添加两条记录 ———— 你是我好友 我是你好友
	} else {

		frqData := []Friend{
			{Appid: appid, Uid: uid, FriendUid: fuid},
			{Appid: appid, Uid: fuid, FriendUid: uid},
		}
		_, err = o.InsertMulti(1, frqData)

	}

	return err
}

/**
  查找申请记录
*/
func FindRecordByFriend(uid int64, fuid int64, appid int64) error {
	o := orm.NewOrm()
	var friend Friend
	err := o.Raw("SELECT appid, uid, friend_uid FROM friend WHERE appid = ? and uid = ? and friend_uid = ?", appid, uid, fuid).QueryRow(&friend)
	return err
}

/**
  删除好友
*/
func DelFriendShip(uid int64, fuid int64, appid int64) error {
	o := orm.NewOrm()
	_,err := o.Raw("DELETE FROM `friend` WHERE `appid` = ? and `uid` = ? and `friend_uid` = ?", appid, uid, fuid).Exec()
	return err
}