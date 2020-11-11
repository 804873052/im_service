package models

import (
	"github.com/astaxie/beego/orm"
)

type GroupMember struct {
	GroupId   int64  `orm:"pk" json:"group_id"`
	Uid       int64  `json:"uid"`
	Timestamp int    `json:"timestamp"`
	Nickname  string `json:"nickname"`
	Mute      int8   `json:"mute"`
	Deleted   int8   `json:"deleted"`
}

//指定表名
func (g *GroupMember) TableName() string {
	return TableName("group_member")
}

// 设置引擎为 INNODB
func (g *GroupMember) TableEngine() string {
	return "INNODB"
}

/**
创建群时添加成员
*/
func MultiInsertGroupMember(groupMemberList []GroupMember, o orm.Ormer) (int64, error) {
	return o.InsertMulti(1, groupMemberList)
}

/**
修改昵称
*/
func (g *GroupMember) ModifyGroupNickName() error {
	o := orm.NewOrm()
	_, err := o.Raw("update `group_member` set `nickname` = ? "+
		"where `group_id` = ? and `uid` = ?", g.Nickname, g.GroupId, g.Uid).Exec()
	return err
}

/**
获取q群成员
*/
func (g *GroupMember) GetGroupMember() (GroupMember, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("group_member")
	groupMember := GroupMember{}
	err := qs.Filter("group_id", g.GroupId).Filter("uid", g.GroupId).One(&groupMember)
	return groupMember, err
}

/**
获取q群成员
*/
func (g *GroupMember) ListGroupMember() ([]GroupMember, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("group_member")
	var groupMemberList []GroupMember
	_, err := qs.Filter("group_id", g.GroupId).Filter("deleted", 0).All(&groupMemberList)
	return groupMemberList, err
}

/**
添加群成员
*/
func (g *GroupMember) AddGroupMember() error {
	o := orm.NewOrm()
	_, err := o.Raw("replace `group_member` values (?, ?, ?, ?, ?, ?)",
		g.GroupId, g.Uid, g.Timestamp, g.Nickname, 0, 0).Exec()
	return err
}

/**
删除群成员
*/
func (g *GroupMember) DeleteGroupMember() error {
	o := orm.NewOrm()
	_, err := o.Raw("update `group_member` set deleted = 1 where `group_id` = ? and `uid` = ?", g.GroupId, g.Uid).Exec()
	return err
}

/**
查看是否存在群成员
*/
func IsGroupMemberExisted(groupId int64, uids []int64) (int64, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("group_member")

	num, err := qs.Filter("group_id", groupId).Filter("uid__in", uids).Filter("deleted", 0).Count()
	return num, err
}
