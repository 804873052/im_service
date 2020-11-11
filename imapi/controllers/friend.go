package controllers

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"imapi/models"
	"imapi/util"
	"strconv"
	"time"
)

////登录后返回的用户信息
//type userInfo struct {
//	Nickname string
//	Avatar string
//	Mobile *string
//	Email string
//	Sex int8
//	Token string
//}
//
//var (
//	UserList map[int]*userInfo
//)

type FriendController struct {
	BaseController
}

/**
1. 添加好友:

v = {
"sender":请求方用户id,
"receiver":接收方用户id,
"content":请求内容,
"timestamp":时间戳(秒)
}
op = {"add_friend_request":v}
*/

type AddFriendRequest struct {
	Sender    int64  `json:"sender"`
	Receiver  int64  `json:"receiver"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

/**
2. 同意/拒绝添加好友:

v = {
"sender":请求方用户id,
"receiver":接收方用户id,
"status":答复状态(0.同意;1.拒绝),
"timestamp":时间戳(秒)
}
op = {"add_friend_reply":v}

*/

type AddFriendReply struct {
	Sender    int64 `json:"sender"`
	Receiver  int64 `json:"receiver"`
	Status    int8  `json:"status"`
	Timestamp int64 `json:"timestamp"`
}

const AGREE_ADD_FRIEND int8 = 0
const REFUSE_ADD_FRIEND int8 = 1

/*
	请求添加好友
*/
// @Title 请求添加好友
// @Description 请求添加好友接口
// @Param	token				header 		string	true		"用户手机号码"
// @Param	friend_uid		    formData 	string	true		"好友用户id"
// @Param	requst_msg		    formData 	string	false		"请求消息"
// @Param	friend_nickname		formData 	string	false		"好友备注"
// @Success 200 "code":0 无错误|非0 出现错误    "msg":"success"|"具体错误内容"  "data":nil
// @router /requestAddFriend [post]
func (c *FriendController) RequestAddFriend() {
	token := c.Ctx.Request.Header.Get("token")

	if "" == token {
		c.Error(util.ERR_MISSPARAM, "token参数错误", nil)
	}

	friendUid, err := c.GetInt64("friend_uid")
	if 0 == friendUid {
		c.Error(util.ERR_MISSPARAM, "好友ID参数错误", nil)
	}

	nickname := c.GetString("friend_nickname","")
	requstMsg := c.GetString("requst_msg","")

	//验证token
	appid, uid, err := AuthToken(token)
	if err != nil {
		c.Error(util.ERR_TOKEN, "token已失效 请重新登录!", nil)
	}

	// 判断该用户是否存在或者是否可以添加
	uAuth := models.UserAuth{} //用户账户结构体
	uAuth.Id = friendUid
	//获取用户注册表信息
	uAuth.GetUserByUid()
	if 0 == uAuth.Appid {
		// 没有该用户
		c.Error(util.ERR_USER_NOTFOUND, "添加用户不存在", nil)
	}

	url := "http://106.53.107.155:6666/post_system_message" + "?appid=" + strconv.FormatInt(appid, 10) + "&uid=" + strconv.FormatInt(friendUid, 10)
	dataMap := make(map[string]AddFriendRequest)
	reqJson := AddFriendRequest{}
	reqJson.Sender = uid
	reqJson.Receiver = friendUid
	reqJson.Content = requstMsg
	reqJson.Timestamp = time.Now().Unix()
	dataMap["add_friend_request"] = reqJson
	req, err := util.Post(dataMap, url)
	if nil != err {
		c.Error(util.ERR_JSON_CONVERT_FAILED, err.Error(), nil)
	}

	//事务处理
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if e := recover(); e != nil {
			o.Rollback()
			logs.Error(e)
			c.Error(util.ERR_TRANS, "申请失败!", nil)
		}
	}()

	// 添加friend_request
	requestId, err := models.InsertRecord(uid, friendUid, appid, &nickname, o)
	if err != nil {
		panic(err)
	}

	// 添加消息
	_, err = models.InsertRecordMessage(requestId, uid, requstMsg, o)
	if err != nil {
		panic(err)
	}
	err = nil

	res, err := req.Response()
	if nil != err {
		panic(err)
	}

	if 200 != res.StatusCode {
		panic("服务器发生异常")
	}

	o.Commit()

	c.Success(nil)
}

/*
	同意添加好友
*/
// @Title 同意添加好友
// @Description 同意添加好友接口
// @Param	token				header 		string	true		"token"
// @Param	request_id		    formData 	int64	true		"好友申请id"
// @Success 200 "code":0 无错误|非0 出现错误    "msg":"success"|"具体错误内容"  "data":nil
// @router /AgreeAddFriend [post]
func (c *FriendController) AgreeAddFriend() {
	token := c.Ctx.Request.Header.Get("token")

	if "" == token {
		c.Error(util.ERR_MISSPARAM, "token参数错误", nil)
	}

	rid, err := c.GetInt64("request_id")
	if 0 == rid {
		c.Error(util.ERR_MISSPARAM, "好友申请ID参数错误", nil)
	}

	//验证token
	appid, uid, err := AuthToken(token)
	if err != nil {
		c.Error(util.ERR_TOKEN, "token已失效 请重新登录!", nil)
	}

	// 查询好友申请记录是否存在
	fr := models.FriendRequest{} //好友申请结构体
	fr.Id = rid
	//获取用户注册表信息
	fr.GetUserByUid()
	if 0 == fr.Appid {
		// 没有该用户
		c.Error(util.ERR_USER_NOTFOUND, "添加记录不存在", nil)
	}
	// 判断该用户是否和记录的用户ID一致
	if uid != fr.FriendUid {
		c.Error(1, "非法操作", nil)
	}
	// 如果是同意状态，则不进行重复提交
	if fr.Status == models.FR_STATUS_AGREE {
		c.Error(util.ERR_REPEAT_APPLY_FRIEND, "请勿重复添加", nil)
	}
	fuid := fr.FriendUid
	// 查询是否已经添加过
	err = models.FindRecordByFriend(fr.Uid, fuid, appid)
	if err == nil {
		c.Error(util.ERR_REPEAT_APPLY_FRIEND, "请勿重复添加", nil)
	}

	url := "http://106.53.107.155:6666/post_system_message" + "?appid=" + strconv.FormatInt(appid, 10) + "&uid=" + strconv.FormatInt(fr.Uid, 10)
	dataMap := make(map[string]AddFriendReply)
	reqJson := AddFriendReply{}
	reqJson.Sender = uid
	reqJson.Receiver = fr.Uid
	reqJson.Status = AGREE_ADD_FRIEND
	reqJson.Timestamp = time.Now().Unix()
	dataMap["add_friend_reply"] = reqJson
	req, err := util.Post(dataMap, url)
	if nil != err {
		c.Error(util.ERR_JSON_CONVERT_FAILED, err.Error(), nil)
	}

	//事务处理
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if e := recover(); e != nil {
			o.Rollback()
			logs.Error(e)
			c.Error(util.ERR_TRANS, "添加失败!", nil)
		}
	}()

	// 添加friend 添加两条
	err = models.InsertRecordFriend(fr.Uid, fuid, appid, o)
	if err != nil {
		panic(err)
	}
	//
	// 修改friend_request
	status := models.FR_STATUS_AGREE
	num, _ := models.UpdateRecordMessage(rid, status, o)
	if num == 0 {
		panic("更新失败")
	}
	err = nil

	res, err := req.Response()
	if nil != err {
		panic(err)
	}

	if 200 != res.StatusCode {
		panic("服务器发生异常")
	}
	o.Commit()
	c.Success(nil)
}

/*
	拒绝添加好友
*/
// @Title 拒绝添加好友
// @Description 拒绝添加好友接口
// @Param	token				header 		string	true		"token"
// @Param	request_id		    formData 	int64	true		"好友申请id"
// @Success 200 "code":0 无错误|非0 出现错误    "msg":"success"|"具体错误内容"  "data":nil
// @router /RefuseAddFriend [post]
func (c *FriendController) RefuseAddFriend() {
	token := c.Ctx.Request.Header.Get("token")

	if "" == token {
		c.Error(util.ERR_MISSPARAM, "token参数错误", nil)
	}

	rid, err := c.GetInt64("request_id")
	if 0 == rid {
		c.Error(util.ERR_MISSPARAM, "好友申请ID参数错误", nil)
	}

	//验证token
	_, uid, err := AuthToken(token)
	if err != nil {
		c.Error(util.ERR_TOKEN, "token已失效 请重新登录!", nil)
	}

	// 查询好友申请记录是否存在
	fr := models.FriendRequest{} //好友申请结构体
	fr.Id = rid
	//获取用户注册表信息
	fr.GetUserByUid()
	if 0 == fr.Appid {
		// 没有该用户
		c.Error(util.ERR_USER_NOTFOUND, "记录不存在", nil)
	}
	// 判断该用户是否和记录的用户ID一致
	if uid != fr.FriendUid {
		c.Error(1, "非法操作", nil)
	}
	// 如果是同意状态，则不进行重复提交
	if fr.Status == models.FR_STATUS_AGREE || fr.Status == models.FR_STATUS_REFUSE {
		c.Error(util.ERR_REPEAT_APPLY_FRIEND, "不可操作", nil)
	}

	//事务处理
	o := orm.NewOrm()

	// 修改friend_request
	status := models.FR_STATUS_REFUSE
	num, _ := models.UpdateRecordMessage(rid, status, o)
	if num == 0 {
		panic("更新失败")
	}
	err = nil
	c.Success(nil)
}

/*
	删除好友
*/
// @Title 删除好友
// @Description 删除好友接口
// @Param	token				header 		string	true		"token"
// @Param	friend_uid		    formData 	string	true		"好友用户id"
// @Success 200 "code":0 无错误|非0 出现错误    "msg":"success"|"具体错误内容"  "data":nil
// @router /DeleteFriend [post]
func (c *FriendController) DeleteFriend() {
	token := c.Ctx.Request.Header.Get("token")
	if "" == token {
		c.Error(util.ERR_MISSPARAM, "token参数错误", nil)
	}

	friendUid,_ := c.GetInt64("friend_uid")
	if 0 == friendUid {
		c.Error(util.ERR_MISSPARAM, "friend_uid参数错误", nil)
	}

	//验证token
	appid, uid, err := AuthToken(token)
	if err != nil {
		c.Error(util.ERR_TOKEN, "token已失效 请重新登录!", nil)
	}

	err = models.DelFriendShip(uid,friendUid,appid)
	if err != nil {
		c.Error(util.ERR_DB, "删除好友失败", nil)
	}

	c.Success(nil)
}

/*
	设置好友备注
*/
// @Title 设置好友备注
// @Description 设置好友备注
// @Param	token				header 		string	true		"token"
// @Param	friend_uid		    formData 	string	true		"好友用户id"
// @Success 200 "code":0 无错误|非0 出现错误    "msg":"success"|"具体错误内容"  "data":nil
// @router /SetFriendNickname [post]
func (c *FriendController) SetFriendNickname() {
	//token := c.Ctx.Request.Header.Get("token")
	////friendUid := c.GetInt64("friend_uid", (int64)0)
	//
	//if ("" == token) {
	//
	//}
	c.Success(nil)
}

/*
	查找好友
*/
// @Title 查找好友
// @Description 查找好友
// @Param	token				header 		string	true		"token"
// @Param	nickname		    formData 		string	true		"查找(昵称)关键字"
// @Success 200 "code":0 无错误|非0 出现错误    "msg":"success"|"具体错误内容"  "data":nil
// @router /ResearchFriend [post]
func (c *FriendController) ResearchFriend() {

	token := c.Ctx.Request.Header.Get("token")

	if token == "" {
		c.Error(util.ERR_MISSPARAM, "token参数不能为空", nil)
	}

	nickname := c.GetString("nickname", "")
	if nickname == "" {
		c.Error(util.ERR_MISSPARAM, "key(昵称)参数不能为空", nil)
	}

	//验证token
	Appid, _, err := AuthToken(token)
	if err != nil {
		c.Error(util.ERR_TOKEN, "token已失效 请重新登录!", nil)
	}

	researchFriends, err := models.SearchByNickname(Appid, nickname)
	if err != nil {
		c.Error(util.ERR_DB, "搜索好友查询失败!", nil)
	}

	c.Success(researchFriends)
}

/*
	好友列表
*/
// @Title 好友列表
// @Description 好友列表
// @Param	token				header 		string	true		"token"
// @Success 200 "code":0 无错误|非0 出现错误    "msg":"success"|"具体错误内容"  "data":nil
// @router /FriendList [post]
func (c *FriendController) FriendList() {
	token := c.Ctx.Request.Header.Get("token")

	if "" == token {
		c.Error(util.ERR_MISSPARAM, "token参数错误", nil)
	}

	//验证token
	appid, uid, err := AuthToken(token)
	if err != nil {
		c.Error(util.ERR_TOKEN, "token已失效 请重新登录!", nil)
	}

	fmt.Printf("appid:%v, uid:%v", appid, uid)

	lst := models.GetMyFriendsList(uid)
	c.Success(lst)
}

/*
	新的朋友列表
*/
// @Title 新的朋友列表
// @Description 新的朋友列表
// @Param	token				header 		string	true		"token"
// @Success 200 "code":0 无错误|非0 出现错误    "msg":"success"|"具体错误内容"  "data":nil
// @router /NewFriendList [post]
func (c *FriendController) NewFriendList() {
	token := c.Ctx.Request.Header.Get("token")

	if "" == token {
		c.Error(util.ERR_MISSPARAM, "token参数错误", nil)
	}

	//验证token
	appid, uid, err := AuthToken(token)
	if err != nil {
		c.Error(util.ERR_TOKEN, "token已失效 请重新登录!", nil)
	}

	fmt.Printf("appid:%v, uid:%v", appid, uid)

	//lst := models.GetMyNewFriendsList(uid)

	lst,err := models.GetMyNewFriendsListV2(uid)
	if err != nil {
		c.Error(util.ERR_DB, err.Error(), nil)
	}

	c.Success(lst)
}

/*
	新的朋友消息
*/
// @Title 新的朋友消息
// @Description 新的朋友消息
// @Param	token				header 		string	true		"token"
// @Param	request_id				formData 	string	true		"好友请求id"
// @Success 200 "code":0 无错误|非0 出现错误    "msg":"success"|"具体错误内容"  "data":nil
// @router /NewFriendMsg [post]
func (c *FriendController) NewFriendMsg() {
	token := c.Ctx.Request.Header.Get("token")

	rid, err := c.GetInt64("request_id")
	if 0 == rid {
		c.Error(util.ERR_MISSPARAM, "好友申请ID参数错误", nil)
	}

	if "" == token {
		c.Error(util.ERR_MISSPARAM, "token参数错误", nil)
	}

	//验证token
	appid, uid, err := AuthToken(token)
	if err != nil {
		c.Error(util.ERR_TOKEN, "token已失效 请重新登录!", nil)
	}

	fmt.Printf("appid:%v, uid:%v", appid, uid)

	lst := models.GetMyNewFriendsMsg(rid)
	c.Success(lst)
}
