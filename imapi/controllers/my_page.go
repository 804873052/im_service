package controllers

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
	"imapi/models"
	"imapi/util"
	"strconv"
)

type MyPageController struct {
	BaseController
}



/*
	设置个性签名
*/
// @Title 设置个性签名
// @Description 设置个性签名接口
// @Param	token		header 	string	true		"用户token"
// @Param	signatrue	formData 	string	true		"用户签名"
// @Success 200 {object} controllers.ReturnMsg
// @router /setSign [post]
func (c *MyPageController) SetSignature() {

	token := c.GetHeaderToken()
	if token == "" {
		c.Error(util.ERR_MISSPARAM,"token参数不能为空",nil)
	}

	signatrue := c.GetString("signatrue","")
	if signatrue == ""{
		c.Error(util.ERR_MISSPARAM,"signatrue参数不能为空",nil)
	}
	//验证token
	Appid,Uid,err := AuthToken(token)
	if err != nil {
		c.Error(util.ERR_TOKEN,"token已失效 请重新登录!",nil)
	}

	uInfo := models.UserInfo{}
	uInfo.Uid = Uid
	uInfo.Appid = Appid
	_,err = uInfo.EditSignatrue(signatrue)
	if err != nil {
		c.Error(util.ERR_DB,err.Error(),nil)
	}
	c.Success(nil)
}

/*
	修改手机号码
*/
// @Title 修改手机号码
// @Description 修改手机号码接口
// @Param	token		header 	string	true		"用户token"
// @Param	new_phone		formData 	string	true	"新的手机号码"
// @Param	captcha		formData 	string	true		"验证码"
// @Success 200 {object} controllers.ReturnMsg
// @router /changePhone [post]
func (c *MyPageController) ChangePhone() {

	token := c.GetHeaderToken()
	if token == "" {
		c.Error(util.ERR_MISSPARAM,"token参数不能为空",nil)
	}
	newPhone := c.GetString("new_phone","")
	if newPhone == "" {
		c.Error(util.ERR_MISSPARAM,"new_phone参数不能为空!",nil)
	}
	captcha := c.GetString("captcha","")
	if  captcha == "" {
		c.Error(util.ERR_MISSPARAM,"captcha参数不能为空!",nil)
	}

	//验证token
	Appid,Uid,err := AuthToken(token)
	fmt.Println(Uid)
	if err != nil {
		c.Error(util.ERR_TOKEN,"token已失效 请重新登录!",nil)
	}

	//短信验证码验证
	conn := models.Redis_pool.Get()
	defer conn.Close()
	key := fmt.Sprintf("phone_%s_%s",strconv.FormatInt(Appid,10),newPhone)
	reply,err := redis.String(conn.Do("GET", key))

	if err != nil {
		c.Error(util.ERR_REDISFAIL,"验证码已过期",nil)
	}

	if reply != captcha {
		c.Error(util.ERR_CAPTCHA,"验证码不匹配！",nil)
	}

	uAuth := models.UserAuth{}//用户账户结构体
	uAuth.Id = Uid
	uAuth.Appid = Appid

	uInfo := models.UserInfo{} //用户信息结构体
	uInfo.Uid = Uid
	uInfo.Appid = Appid

	//事务处理
	o := orm.NewOrm()
	defer func() {
		if e := recover(); e != nil {
			o.Rollback()
			logs.Error(e)
			c.Error(util.ERR_TRANS,"修改手机号码失败!",nil)
		}
	}()
	o.Begin()
	//更新user_auth表
	qs1 := o.QueryTable(&uAuth)
	_,err = qs1.Filter("id", uAuth.Id).Filter("Appid", uAuth.Appid).Update(orm.Params{
		"phone": newPhone,
	})

	if err != nil {
		panic(err)
	}
	//更新user_info表
	qs2 := o.QueryTable(&uInfo)
	_,err = qs2.Filter("uid", uInfo.Uid).Filter("Appid", uInfo.Appid).Update(orm.Params{
		"phone": newPhone,
	})

	if err != nil {
		panic(err)
	}

	err = nil
	o.Commit()
	c.Success(nil)

}

/*
	手机号码是否已被注册
*/
// @Title 手机号码是否已被注册
// @Description 手机号码是否已被注册接口
// @Param	token		header 	string	true		"用户token"
// @Param	phone		formData 	string	true	"手机号码"
// @Success 200 {object} controllers.ReturnMsg
// @router /phoneIsAvialble [post]
func (c *MyPageController) PhoneIsAvialble() {

	token := c.GetHeaderToken()
	if token == "" {
		c.Error(util.ERR_MISSPARAM,"token参数不能为空",nil)
	}
	phone := c.GetString("phone","")
	if phone == "" {
		c.Error(util.ERR_MISSPARAM,"phone参数不能为空",nil)
	}
	//验证token
	Appid,_,err := AuthToken(token)
	if err != nil {
		c.Error(util.ERR_TOKEN,"token已失效 请重新登录!",nil)
	}

	uAuth := models.UserAuth{}
	uAuth.Phone = phone
	uAuth.Appid = Appid
	isReg,_ := uAuth.GetUserByPhone()

	if err != nil {
		c.Error(util.ERR_DB,err.Error(),nil)
	}

	ret := map[string]string{
		"isReg":strconv.FormatInt(isReg,10) ,
	}

	c.Success(ret)
}

/*
	用户反馈
*/
// @Title 用户反馈
// @Description 用户反馈接口
// @Param	token		header 	string	true		"用户token"
// @Param	phone		formData 	string	true	"手机号码"
// @Param	content  	formData 	string	true	"用户反馈"
// @Success 200 {object} controllers.ReturnMsg
// @router /feedback [post]
func (c *MyPageController) Feedback() {

	token := c.GetHeaderToken()
	if token == "" {
		c.Error(util.ERR_MISSPARAM,"token参数不能为空",nil)
	}
	content := c.GetString("content","")
	if content == "" {
		c.Error(util.ERR_MISSPARAM,"content参数不能为空",nil)
	}
	phone := c.GetString("phone","")
	if phone == "" {
		c.Error(util.ERR_MISSPARAM,"phone参数不能为空",nil)
	}
	//验证token
	Appid,Uid,err := AuthToken(token)
	if err != nil {
		c.Error(util.ERR_TOKEN,"token已失效 请重新登录!",nil)
	}

	feedback := models.Feedback{}
	feedback.Uid = Uid
	feedback.Appid = Appid
	feedback.Phone = phone
	feedback.Content = content

	o := orm.NewOrm()
	_,err = o.Insert(&feedback)

	if err != nil {
		c.Error(util.ERR_DB,err.Error(),nil)
	}

	c.Success(nil)

}


/*
	更新我的用户信息
*/
// @Title 更新我的用户信息
// @Description 更新我的用户信息接口
// @Param	token		header 	string	true		"用户token"
// @Param	headimg	    formData 	string	false		"用户头像"
// @Param	nickname		formData 	string	false		"昵称"
// @Param	region_id		formData 	string	false   "地区主键"
// @Param	sex		formData 	string	false		"性别：0.未设置；1.男；2.女；3.保密"
// @Param	signature	formData 	string	false		"用户签名"
// @Success 200 {object} controllers.MyInfo
// @router /updateMyInfo [post]
func (c *MyPageController) UpdateMyInfo() {

	token := c.GetHeaderToken()
	if token == "" {
		c.Error(util.ERR_MISSPARAM, "token参数不能为空", nil)
	}
	//验证token
	_, Uid, err := AuthToken(token)
	if err != nil {
		c.Error(util.ERR_TOKEN, "token已失效 请重新登录!", nil)
	}

	uAuth := models.UserAuth{} //用户账户结构体
	uAuth.Id = Uid
	//获取用户注册表信息
	uAuth.GetUserByUid()

	if uAuth.Phone == "" || uAuth.Appid == 0 {
		c.Error(util.ERR_NOFOUND, "无注册信息!", nil)
	}

	nickname := c.GetString("nickname")
	sex,_ := c.GetInt8("sex")
	headimg := c.GetString("headimg")
	regionId,_ := c.GetInt("region_id")
	signature := c.GetString("signature")

	//更新用户信息
	_,err = models.UpdateMyInfo(Uid,nickname,sex,headimg,regionId,signature)

	if err != nil {
		c.Error(util.ERR_DB, err.Error(), nil)
	}

	//更新后重新获取用户信息
	uInfo := models.UserInfo{} //用户信息结构体
	//获取参数信息
	uInfo.Uid = uAuth.Id
	uInfo.GetUserByUid()

	MyInfo := MyInfo{}
	MyInfo.Uid = uInfo.Uid
	MyInfo.Token = token
	MyInfo.Nickname = uInfo.Nickname
	MyInfo.HeadImg = uInfo.HeadImg
	MyInfo.Phone = uInfo.Phone
	MyInfo.Sex = uInfo.Sex
	MyInfo.Signature = uInfo.Signature
	MyInfo.RegionId = uInfo.RegionId

	var regionName string

	if uInfo.RegionId != 0 {
		regionName,_ = models.GetRegoinName(uInfo.RegionId)
	} else {
		regionName = ""
	}

	MyInfo.RegionName = regionName

	//内存存用户列表
	UserList[uInfo.Uid] = &MyInfo

	c.Success(MyInfo)

}