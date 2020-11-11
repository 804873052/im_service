package models

import (
	"github.com/astaxie/beego/orm"
)

type Region struct {
	Id int `json:"Id"`
	Pid int `json:"pid"`
	Path string `json:"path"`
	Level int `json:"level"`
	Name string `json:"name"`
	Code string `json:"code"`
}

//指定表名
func (u *Region) TableName() string {
	return TableName("region")
}

type TreeList struct {
	Id   int `json:"id"`
	RegionName string `json:"region_name"`
	Pid int `json:"pid"`
	Path string `json:"path"`
	Level int `json:"level"`
	Code string `json:"code"`
	Children []*TreeList `json:"children"`
}


var RegionTreeList []*TreeList

func InitTreeList() []*TreeList {
	//读数据库
	if RegionTreeList == nil {
		regionList := []Region{}
		o:=orm.NewOrm()
		qs := o.QueryTable("region")
		qs.All(&regionList)
		//赋值给缓存
		RegionTreeList = NewGetTreeList(regionList,1)
		return RegionTreeList
	//读内存缓存
	} else {
		return RegionTreeList
	}
}


//反复查询数据库获取分类树结果
func GetTreeList(pid int) []*TreeList{
	//获取所有的分类数据
	o := orm.NewOrm()
	var Region []Region
	_,_ = o.QueryTable("region").Filter("pid", pid).All(&Region)
	treeList := []*TreeList{}
	for _, v := range Region{
		child := GetTreeList(v.Id)
		node := &TreeList{
			Id:v.Id,
			RegionName:v.Name,
			Pid:v.Pid,
			Path:v.Path,
			Code:v.Code,
			Level:v.Level,
		}
		node.Children = child
		treeList = append(treeList, node)
	}
	return treeList
}


//public static function listToTreeMulti($list, $root = 0, $pk = 'id', $pid = 'parentId', $child = 'child')


//反复切片获取分类树结果
func NewGetTreeList(list []Region,root int) []*TreeList{
	//获取所有的分类数据
	treeList := []*TreeList{}
	for _, v := range list{
		if(v.Pid == root){
			child := NewGetTreeList(list,v.Id)
			node := &TreeList{
				Id:v.Id,
				RegionName:v.Name,
				Pid:v.Pid,
				Path:v.Path,
				Code:v.Code,
				Level:v.Level,
			}
			node.Children = child
			treeList = append(treeList, node)
		}
	}
	return treeList
}


