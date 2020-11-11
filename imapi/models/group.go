package models

import "github.com/astaxie/beego/orm"

type Group struct {
	Id      int64  `orm:"pk" json:"id"`
	Appid   int64  `json:"appid"`
	Master  int64  `json:"master"`
	Super   int8   `json:"super"`
	Name    string `json:"name"`
	Notice  string `json:"notice"`
	Deleted int8   `json:"deleted"`
}

//指定表名
func (g *Group) TableName() string {
	return TableName("group")
}

// 设置引擎为 INNODB
func (g *Group) TableEngine() string {
	return "INNODB"
}

/**
创建群
*/
func InsertGroup(group Group, o orm.Ormer) (int64, error) {
	return o.Insert(&group)
}

/**
获取群信息
*/
func GetGroupInfo(groupId int64) (Group, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("group")
	group := Group{}
	err := qs.Filter("id", groupId).One(&group)
	return group, err
}

/**
修改群信息
*/
func ModifyGroupInfo(group Group) error {
	o := orm.NewOrm()
	_, err := o.Update(&group)
	return err
}
