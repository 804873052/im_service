package controllers

import (
	"github.com/astaxie/beego/logs"
	"github.com/axgle/pinyin"
	"imapi/models"
	"imapi/util"
	"sort"
	"unicode"
)

type ContactListController struct {
	BaseController
}

type contacter struct {
	Uid int64 `json:"uid"`
	Headimg string `json:"headimg"`
	NickName string `json:"nickname"`
	Pinyin string `json:"pinyin"`
}

//重写通讯录切片的排序规则 以拼音字母的升序排序
type contacterSlice []contacter
func (s contacterSlice) Len() int { return len(s) }
func (s contacterSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s contacterSlice) Less(i, j int) bool { return s[i].Pinyin < s[j].Pinyin }

/*
	通讯录列表
*/
// @Title 通讯录列表
// @Description 联系人结构体 contacter struct { Uid int64 Headimg string NickName string Pinyin string }
// @Param	token		header 	string	true		"用户token"
// @Success 200 {object} controllers.ReturnMsg
// @router /contactList [post]
func (c *ContactListController) ContactList() {

	token := c.GetHeaderToken()
	if token == ""  {
		c.Error(util.ERR_MISSPARAM, "缺少token参数!", nil)
	}

	//验证token
	_,Uid,err := AuthToken(token)
	if err != nil {
		c.Error(util.ERR_TOKEN,"token已失效 请重新登录!",nil)
	}

	//获取我的好友列表
	contactList,err := models.GetMyContactList(Uid)
	if err != nil {
		c.Error(0,err.Error(),nil)
	}

	contactListSlice := contacterSlice{}
	//赋值通讯录切片
	for _,v := range contactList {
		pinYin := ""
		//判断v.Nickname第一个字符是否为数字或符号 是的话给pinyin标志"#"
		firstChar := []rune(v.Nickname[:1])
		//unicode.IsDigit 是数字返回真 || unicode.IsPunct 为Unicode标点字符返回真
		if unicode.IsDigit(firstChar[0]) || unicode.IsPunct(firstChar[0]) {
			pinYin = "#"
		} else {
			pinYin = pinyin.Convert(v.Nickname)
		}
		contactListSlice = append(contactListSlice,contacter{
			Uid: v.Uid,
			Headimg: v.HeadImg,
			NickName: v.Nickname,
			Pinyin: pinYin,
		})
	}
	//按拼音字母的升序稳定排序后返回
	sort.Stable(contactListSlice)
	c.Success(contactListSlice)

}


/*
	通讯录个人信息
*/
// @Title 通讯录个人信息
// @Description 返回切片字段 {uid int64,nickname string,headImg string,signature,string,regionId int,regionName string}
// @Param	token		header 	string	true		"用户token"
// @Param	f_uid		formData 	string	true	"联系人的主键id"
// @Success 200 {object} controllers.ReturnMsg
// @router /contacterInfo [post]
func (c *ContactListController) ContacterInfo() {

	token := c.GetHeaderToken()
	if token == ""  {
		c.Error(util.ERR_MISSPARAM, "缺少token参数!", nil)
	}

	friendUid,_ := c.GetInt64("f_uid", 0)
	if friendUid == 0  {
		c.Error(util.ERR_MISSPARAM, "缺少f_uid参数!", nil)
	}

	//验证token
	_,myId,err := AuthToken(token)
	if err != nil {
		c.Error(util.ERR_TOKEN,"token已失效 请重新登录!",nil)
	}
	//判断是否为我的好友
	num,err := models.IsMyFriend(myId,friendUid)
	if err != nil {
		c.Error(util.ERR_DB,err.Error(),nil)
	}

	oneUser := models.UserInfo{}
	oneUser.Uid = friendUid
	oneUser.GetUserByUid()

	if oneUser.Phone == "" {
		c.Error(util.ERR_NOFOUND,"没有查询到该好友信息",nil)
	}

	var regionName string

	if oneUser.RegionId != 0 {
		regionName,_ = models.GetRegoinName(oneUser.RegionId)
	} else {
		regionName = ""
	}

	ret := map[string]interface{}{
		"uid" : oneUser.Uid,
		"nickname" : oneUser.Nickname,
		"headimg" : oneUser.HeadImg,
		"signature" : oneUser.Signature,
		"region_id" : oneUser.RegionId,
		"region_name" : regionName,
		"sex":oneUser.Sex,
		"phone":oneUser.Phone,
		"is_my_friend":num,
	}

	c.Success(ret)

}


/*
	设置备注与标签
*/
// @Title 设置备注与标签
// @Description 返回切片字段 {uid int64,nickname string,headImg string,signature,string,regionId int,regionName string}
// @Param	token		header 	string	true		"用户token"
// @Param	f_uid		formData 	string	true	"联系人的主键id"
// @Param	remarks		formData 	string	false	"备注信息"
// @Param	tags		formData 	string	false	"标签信息"
// @Success 200 {object} controllers.ReturnMsg
// @router /setRemarksAndTags [post]
func (c *ContactListController) SetRemarksAndTags() {

	token := c.GetHeaderToken()
	if token == ""  {
		c.Error(util.ERR_MISSPARAM, "缺少token参数!", nil)
	}

	friendUid,_ := c.GetInt64("f_uid", 0)
	if friendUid == 0  {
		c.Error(util.ERR_MISSPARAM, "缺少f_uid参数!", nil)
	}

	//验证token
	_,_,err := AuthToken(token)
	if err != nil {
		c.Error(util.ERR_TOKEN,"token已失效 请重新登录!",nil)
	}

	remarks := c.GetString("remarks","")
	tags := c.GetString("tags","")
	logs.Info(remarks,tags)
}
