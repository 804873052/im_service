package controllers

import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"imapi/models"
	"imapi/util"
	"strconv"
	"strings"
	"time"
)

type GroupController struct {
	BaseController
}

func (c *GroupController) listUserByUids(appid int64, uid int64) []models.UserInfo {
	uids := c.GetString("uids")
	var uidList []int64
	if "" == uids {
		uidList = make([]int64, 0)
		uidList = append(uidList, uid)
	} else {
		uidArray := strings.Split(uids, ",")
		uidList = make([]int64, 0)
		uidList = append(uidList, uid)
		for i := range uidArray {
			if "" == uidArray[i] {
				c.Error(util.ERR_MISSPARAM, "uids中不能存在空值", nil)
			}

			tmpUid, err := strconv.ParseInt(uidArray[i], 10, 64)
			if nil != err {
				c.Error(util.ERR_TYPECONVERSION, err.Error(), nil)
			}
			uidList = append(uidList, tmpUid)
		}
	}

	userInfoList, err := models.GetUserByUid(appid, uidList)
	if nil != err {
		c.Error(util.ERR_DB, err.Error(), nil)
	}

	if len(userInfoList) != len(uidList) {
		c.Error(util.ERR_MISSPARAM, "uids用户不存在", nil)
	}

	return userInfoList
}

func (c *GroupController) getAuth() (int64, int64) {
	token := c.Ctx.Request.Header.Get("token")
	if "" == token {
		c.Error(util.ERR_MISSPARAM, "请求头部token不存在", nil)
	}

	//验证token
	appid, uid, err := AuthToken(token)
	if err != nil {
		c.Error(util.ERR_TOKEN, "token已失效 请重新登录!", nil)
	}
	return appid, uid
}

/*
	拉人入群
*/
// @Title 拉人入群
// @Description 拉人入群
// @Param	token				header 		string	true		"token"
// @Param	uids		    	formData 		string	true	"多个好友"
// @Success 200 "code":0 无错误|非0 出现错误    "msg":"success"|"具体错误内容"  "data":nil
// @router /CreateGroup [post]
func (c *GroupController) CreateGroup() {
	appid, uid := c.getAuth()

	userInfoList := c.listUserByUids(appid, uid)
	groupName := ""
	for i := range userInfoList {
		if userInfoList[i].Uid == uid {
			groupName = userInfoList[i].Nickname + groupName
		} else {
			groupName += "," + userInfoList[i].Nickname
		}
	}

	//事务处理
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if e := recover(); e != nil {
			o.Rollback()
			logs.Error(e)
			c.Error(util.ERR_TRANS, "创建群失败!", nil)
		}
	}()

	group := models.Group{}
	group.Appid = appid
	group.Master = uid
	group.Name = groupName
	group.Super = 0

	// 添加friend_request
	groupId, err := models.InsertGroup(group, o)
	if err != nil {
		panic(err)
	}

	memberNum := len(userInfoList)
	groupMemberList := make([]models.GroupMember, memberNum)
	for i := range userInfoList {
		groupMember := models.GroupMember{}
		groupMember.Uid = userInfoList[i].Uid
		groupMember.GroupId = groupId
		groupMember.Mute = 0
		groupMember.Deleted = 0
		groupMember.Timestamp = int(time.Now().Unix())
		groupMember.Nickname = userInfoList[i].Nickname
		groupMemberList[i] = groupMember
	}

	// 添加消息
	_, err = models.MultiInsertGroupMember(groupMemberList, o)
	if err != nil {
		panic(err)
	}
	err = nil
	o.Commit()
	c.Success(groupId)
}

func (c *GroupController) getGroupInfoByGroupId() (models.Group, int64) {

	groupId, err := c.GetInt64("group_id", 0)
	if err != nil {
		c.Error(util.ERR_MISSPARAM, "groupId参数错误!", nil)
	}

	//验证token
	_, uid := c.getAuth()

	group, err := models.GetGroupInfo(groupId)
	if err != nil {
		c.Error(util.ERR_DB, err.Error(), nil)
	}

	return group, uid
}

/*
	获取群信息
*/
// @Title 获取群信息
// @Description 获取群信息
// @Param	token				header 			string	true		"token"
// @Param	group_id			query 			string	true		"群组id"
// @Success 200 "code":0 无错误|非0 出现错误    "msg":"success"|"具体错误内容"  "data":nil
// @router /GetGroupInfo [get]
func (c *GroupController) GetGroupInfo() {
	group, _ := c.getGroupInfoByGroupId()
	c.Success(group)
}

/*
	修改群名
*/
// @Title 修改群名
// @Description 修改群名
// @Param	token				header 			string	true		"token"
// @Param	group_id			formData 		string	true		"群组id"
// @Param	name		    	formData 		string	true		"群组名称"
// @Success 200 "code":0 无错误|非0 出现错误    "msg":"success"|"具体错误内容"  "data":nil
// @router /ModifyGroupName [post]
func (c *GroupController) ModifyGroupName() {
	name := c.GetString("name")
	if "" == name {
		c.Error(util.ERR_MISSPARAM, "name参数错误", nil)
	}
	group, uid := c.getGroupInfoByGroupId()

	if uid != group.Master {
		c.Error(util.ERR_NOT_MASTER, "抱歉，您不是群主，无法修改群名", nil)
	}

	group.Name = name
	err := models.ModifyGroupInfo(group)
	if err != nil {
		c.Error(util.ERR_DB, err.Error(), nil)
	}
	c.Success(nil)
}

/*
	修改群公告
*/
// @Description 修改群公告
// @Param	token				header 			string	true		"token"
// @Param	group_id			formData 		string	true		"群组id"
// @Param	notice			    formData 		string	true		"群公告"
// @Success 200 "code":0 无错误|非0 出现错误    "msg":"success"|"具体错误内容"  "data":nil
// @router /ModifyGroupNotice [post]
func (c *GroupController) ModifyGroupNotice() {
	notice := c.GetString("notice")
	if "" == notice {
		c.Error(util.ERR_MISSPARAM, "notice参数错误", nil)
	}
	group, uid := c.getGroupInfoByGroupId()

	if uid != group.Master {
		c.Error(util.ERR_NOT_MASTER, "抱歉，您不是群主，无法修改群公告", nil)
	}

	group.Notice = notice
	err := models.ModifyGroupInfo(group)
	if err != nil {
		c.Error(util.ERR_DB, err.Error(), nil)
	}
	c.Success(nil)
}

/*
	修改自己在群内的昵称
*/
// @Description 修改自己在群内的昵称
// @Param	token				header 			string	true		"token"
// @Param	group_id			formData 		string	true		"群组id"
// @Param	nickname		    formData 		string	true		"自己在群内的昵称"
// @Success 200 "code":0 无错误|非0 出现错误    "msg":"success"|"具体错误内容"  "data":nil
// @router /ModifyGroupNickName [post]
func (c *GroupController) ModifyGroupNickName() {
	nickname := c.GetString("nickname")
	if "" == nickname {
		c.Error(util.ERR_MISSPARAM, "nickname参数错误", nil)
	}

	groupId, err := c.GetInt64("group_id", 0)
	if err != nil {
		c.Error(util.ERR_MISSPARAM, "groupId参数错误!", nil)
	}

	_, uid := c.getAuth()

	groupMember := models.GroupMember{}
	groupMember.GroupId = groupId
	groupMember.Uid = uid
	groupMember.Nickname = nickname

	err = groupMember.ModifyGroupNickName()
	if err != nil {
		c.Error(util.ERR_DB, err.Error(), nil)
	}
	c.Success(nil)
}

/*
	获取群成员
*/
// @Title 获取群成员
// @Description 获取群成员
// @Param	token				header 			string	true		"token"
// @Param	group_id			query 		string	true		"群组id"
// @Success 200 "code":0 无错误|非0 出现错误    "msg":"success"|"具体错误内容"  "data":nil
// @router /ListGroupMember [get]
func (c *GroupController) ListGroupMember() {
	groupId, err := c.GetInt64("group_id", 0)
	if err != nil {
		c.Error(util.ERR_MISSPARAM, "groupId参数错误!", nil)
	}

	c.getAuth()
	groupMember := models.GroupMember{}
	groupMember.GroupId = groupId
	groupMemberList, err := groupMember.ListGroupMember()
	if err != nil {
		c.Error(util.ERR_DB, err.Error(), nil)
	}
	c.Success(groupMemberList)
}

func (c *GroupController) getUidListByUids() []int64 {
	uids := c.GetString("uids")
	uidArray := strings.Split(uids, ",")
	uidList := make([]int64, 0)
	for i := range uidArray {
		if "" == uidArray[i] {
			c.Error(util.ERR_MISSPARAM, "uids中不能存在空值", nil)
		}

		tmpUid, err := strconv.ParseInt(uidArray[i], 10, 64)
		if nil != err {
			c.Error(util.ERR_TYPECONVERSION, err.Error(), nil)
		}
		uidList = append(uidList, tmpUid)
	}
	return uidList
}

/*
	添加群成员
*/
// @Title 添加群成员
// @Description 添加群成员
// @Param	token				header 			string	true		"token"
// @Param	group_id			formData 		string	true		"群组id"
// @Param	uid					formData 		string	true		"用户id"
// @Success 200 "code":0 无错误|非0 出现错误    "msg":"success"|"具体错误内容"  "data":nil
// @router /AddGroupMember [post]
func (c *GroupController) AddGroupMember() {

	groupUid, err := c.GetInt64("uid", 0)
	if err != nil {
		c.Error(util.ERR_MISSPARAM, "groupId参数错误!", nil)
	}

	group, uid := c.getGroupInfoByGroupId()

	if uid != group.Master {
		c.Error(util.ERR_NOT_MASTER, "抱歉，您不是群主，无法添加群成员", nil)
	}

	groupUids := make([]int64, 0)
	groupUids = append(groupUids, groupUid)
	num, err := models.IsGroupMemberExisted(group.Id, groupUids)
	if nil != err {
		c.Error(util.ERR_DB, err.Error(), nil)
	}

	if 0 != num {
		c.Error(util.ERR_REPEAT_ADD_GROUP_MEMBER, "抱歉，不能重复添加群成员", nil)
	}

	UserInfoList, err := models.GetUserByUid(group.Appid, groupUids)
	if nil != err {
		c.Error(util.ERR_DB, err.Error(), nil)
	}

	if 0 == len(UserInfoList) {
		c.Error(util.ERR_NOTFOUND, "不存在该用户", nil)
	}

	groupMember := models.GroupMember{}
	groupMember.GroupId = group.Id
	groupMember.Uid = groupUid
	groupMember.Nickname = UserInfoList[0].Nickname
	groupMember.Mute = 0
	groupMember.Deleted = 0
	groupMember.Timestamp = int(time.Now().Unix())

	err = groupMember.AddGroupMember()
	if nil != err {
		c.Error(util.ERR_DB, err.Error(), nil)
	}
	c.Success(nil)
}

/*
	删除群成员
*/
// @Title 删除群成员
// @Description 删除群成员
// @Param	token				header 			string	true		"token"
// @Param	group_id			formData 		string	true		"群组id"
// @Param	uid					formData 		string	true	"用户id"
// @Success 200 "code":0 无错误|非0 出现错误    "msg":"success"|"具体错误内容"  "data":nil
// @router /DeleteGroupMember [post]
func (c *GroupController) DeleteGroupMember() {
	uid, err := c.GetInt64("uid", 0)
	if err != nil {
		c.Error(util.ERR_MISSPARAM, "groupId参数错误!", nil)
	}

	group, master := c.getGroupInfoByGroupId()

	if master != group.Master {
		c.Error(util.ERR_NOT_MASTER, "抱歉，您不是群主，无法删除群成员", nil)
	}

	groupMember := models.GroupMember{}
	groupMember.GroupId = group.Id
	groupMember.Uid = uid
	err = groupMember.DeleteGroupMember()
	if err != nil {
		c.Error(util.ERR_DB, err.Error(), nil)
	}
	c.Success(nil)
}
