package controllers

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
	"imapi/models"
	"imapi/util"
	"strconv"
	"time"
)

const CHARSET = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

//登录后返回的用户信息
type MyInfo struct {
	Uid        int64  `json:"uid"`
	Nickname   string `json:"nickname"`
	Phone      string `json:"phone"`
	HeadImg    string `json:"headimg"`
	Sex        int8   `json:"sex"`
	Signature  string `json:"signature"`
	RegionId   int    `json:"region_id"`
	RegionName string `json:"region_name"`
	Token      string `json:"token"`
}

/**
3. 用户踢下线:

        v = {
            "sender":请求方用户id,
            "timestamp":时间戳(秒)
        }

        op = {"kick_user":v}

*/

type KickUser struct {
	Token     string `json:"token"`
	Sender    int64  `json:"sender"`
	Timestamp int64  `json:"timestamp"`
}

var (
	UserList map[int64]*MyInfo
)

type UserController struct {
	BaseController
}

func init() {
	UserList = make(map[int64]*MyInfo)
}

// @Title 测试接口
// @Description 测试接口
// @Success 200
// @router /test [get]
func (c *UserController) Test() {
	a := c.Ctx.Request.FormValue("nickname")
	fmt.Println(a)
	nickname := c.GetString("nickname")

	fmt.Println(nickname)
	c.Success(nickname)
}

/*
	注册逻辑
*/
// @Title 用户注册
// @Description 用户注册接口
// @Param	phone		formData 	string	true		"用户手机号码"
// @Param	appid		formData 	string	true		"产品id"
// @Param	captcha		formData 	string	true		"短信验证码"
// @Success 200 {object} controllers.ReturnMsg
// @router /reg [post]
func (c *UserController) Register() {

	phone := c.GetString("phone", "")

	if phone == "" {
		c.Error(util.ERR_MISSPARAM, "phone参数不能为空!", nil)
	}

	appid := c.GetString("appid", "")

	if appid == "" {
		c.Error(util.ERR_MISSPARAM, "appid参数不能为空!", nil)
	}

	captcha := c.GetString("captcha", "")

	if captcha == "" {
		c.Error(util.ERR_MISSPARAM, "captcha参数不能为空!", nil)
	}

	//短信验证码验证
	conn := models.Redis_pool.Get()
	defer conn.Close()

	// 默认 captcha:123456 不需要验证
	key := fmt.Sprintf("phone_%s_%s", appid, phone)
	reply, err := redis.String(conn.Do("GET", key))

	if err != nil {
		c.Error(util.ERR_REDISFAIL, "验证码已过期", nil)
	}

	if reply != captcha {
		c.Error(util.ERR_CAPTCHA, "验证码不匹配！", nil)
	}



	uAuth := models.UserAuth{} //用户账户结构体
	uAuth.Phone = c.GetString("phone", "")
	uAuth.Appid, _ = c.GetInt64("appid", 0)
	uAuth.Password = c.GetString("password", "")
	uAuth.WxCode = c.GetString("qq_code", "")
	uAuth.QqCode = c.GetString("wx_code", "")

	//isReg, err := uAuth.GetUserByPhone()
	//if err != nil {
	//	c.Error(util.ERR_DB, err.Error(), nil)
	//}
	//
	//if isReg == 1 {
	//	c.Error(util.ERR_REPLYREG, "重复注册", nil)
	//}

	o := orm.NewOrm()

	// 三个返回参数依次为：是否新创建的，对象 Id 值，错误
	_, id, err := o.ReadOrCreate(&uAuth, "phone", "appid")
	if err != nil {
		c.Error(util.ERR_DB, err.Error(), nil)
	}
	//id, err := o.Insert(&uAuth)
	token, _, err := SetAuthToken(conn, uAuth.Appid, id)
	if err != nil {
		c.Error(util.ERR_REDISFAIL, "imToken写入redis失败", nil)
	}

	c.Success(token)

}

/*
	完善注册信息逻辑(注册流程)
*/
// @Title 完善注册信息
// @Description 完善注册信息接口
// @Param	token		header 	string	true		"用户token"
// @Param	headimg	    formData 	string	false		"用户头像"
// @Param	nickname		formData 	string	true		"昵称"
// @Param	region_id		formData 	string	false   "地区主键"
// @Param	sex		formData 	string	false		"性别：0.未设置；1.男；2.女；3.保密"
// @Success 200 {object} controllers.MyInfo
// @router /setInfo [post]
func (c *UserController) SetUserInfo() {

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

	uInfo := models.UserInfo{} //用户信息结构体
	//获取参数信息
	uInfo.Uid = uAuth.Id
	uInfo.Appid = uAuth.Appid
	uInfo.Phone = uAuth.Phone
	if uAuth.Phone == "" || uAuth.Appid == 0 {
		c.Error(util.ERR_NOFOUND, "无注册信息!", nil)
	}

	uInfo.Nickname = c.GetString("nickname", "")
	if uInfo.Nickname == "" {
		c.Error(util.ERR_MISSPARAM, "nickname参数不能为空!", nil)
	}

	uInfo.Sex, _ = c.GetInt8("sex", 0)
	uInfo.HeadImg = c.GetString("headimg", "")
	uInfo.RegionId, _ = c.GetInt("region_id", 0)

	isReg, err := uInfo.IsReg()
	if err != nil {
		c.Error(util.ERR_DB, err.Error(), nil)
	}

	if isReg == 1 {
		c.Error(util.ERR_REPLYREG, "用户信息重复注册", nil)
	}

	o := orm.NewOrm()
	_, err = o.Insert(&uInfo)

	if err != nil {
		c.Error(util.ERR_DB, err.Error(), nil)
	}

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
		regionName, _ = models.GetRegoinName(uInfo.RegionId)
	} else {
		regionName = ""
	}

	MyInfo.RegionName = regionName

	//内存存用户列表
	UserList[uInfo.Uid] = &MyInfo

	c.Success(MyInfo)

}

/*
	发送短信验证码功能
*/
// @Title 发送短信验证码
// @Description 短信验证码接口
// @Param	phone		formData 	string	true		"用户手机号码"
// @Param	appid		formData 	string	true		"产品id"
// @Success 200 {string} controllers.ReturnMsg { "code": 0, "msg": "success", "data": { "captcha": "339993", "isReg": "1" } }
// @router /sendSms [post]
func (c *UserController) SendSms() {

	phone := c.GetString("phone", "")

	if phone == "" {
		c.Error(util.ERR_MISSPARAM, "phone参数不能为空!", nil)
	}

	appid := c.GetString("appid", "")

	if appid == "" {
		c.Error(util.ERR_MISSPARAM, "appid参数不能为空!", nil)
	}

	var Captcha string = "123456"
	var err error

	if beego.AppConfig.String("runmode") != "dev" {
		Captcha, err = util.SendSmsCaptcha(phone)
		if err != nil {
			c.Error(util.ERR_SYSTEM, err.Error(), nil)
		}
	}

	//短信验证码存redis
	conn := models.Redis_pool.Get()
	defer conn.Close()
	key := fmt.Sprintf("phone_%s_%s", appid, phone)
	_, err = conn.Do("SET", key, Captcha)
	conn.Do("PEXPIRE", key, 1000*60*5) //缓存5分钟

	if err != nil {
		c.Error(util.ERR_REDISFAIL, err.Error(), nil)
	}

	//判断是登录流程还是注册流程 0 注册 1 登录
	uInfo := models.UserInfo{} //用户信息结构体
	//获取参数信息
	uInfo.Appid, _ = strconv.ParseInt(appid, 10, 64)
	uInfo.Phone = phone
	isReg, err := uInfo.IsReg()
	if err != nil {
		c.Error(util.ERR_DB, err.Error(), nil)
	}

	ret := map[string]string{
		"isReg": strconv.FormatInt(isReg, 10),
	}
	c.Success(ret)
}

/*
	登录逻辑
*/
// @Title 登录逻辑
// @Description 登录接口
// @Param	phone		formData 	string	true		"用户手机号码"
// @Param	appid		formData 	string	true		"产品id"
// @Param	captcha		formData 	string	true		"短信验证码"
// @Success  200  {object}  controllers.MyInfo
// @router /login [post]
func (c *UserController) Login() {

	phone := c.GetString("phone", "")
	if phone == "" {
		c.Error(util.ERR_MISSPARAM, "phone参数不能为空!", nil)
	}
	appid := c.GetString("appid", "")
	if appid == "" {
		c.Error(util.ERR_MISSPARAM, "appid参数不能为空!", nil)
	}
	captcha := c.GetString("captcha", "")
	if captcha == "" {
		c.Error(util.ERR_MISSPARAM, "captcha参数不能为空!", nil)
	}

	//短信验证码存redis
	conn := models.Redis_pool.Get()
	defer conn.Close()

	// 默认 captcha:123456 不需要验证
	key := fmt.Sprintf("phone_%s_%s", appid, phone)
	reply, err := redis.String(conn.Do("GET", key))

	if err != nil {
		c.Error(util.ERR_REDISFAIL, "验证码已过期", nil)
	}

	if reply != captcha {
		c.Error(util.ERR_CAPTCHA, "验证码不匹配！", nil)
	}



	uInfo := models.UserInfo{}
	uInfo.Phone = phone
	uInfo.Appid, _ = strconv.ParseInt(appid, 10, 64)

	oneUser, err := uInfo.GetUserInfoByPhone()
	if err != nil {
		c.Error(util.ERR_DB, "未查询到注册信息!", nil)
	}

	token, oldToken, err := SetAuthToken(conn, oneUser.Appid, oneUser.Uid)
	if err != nil {
		c.Error(util.ERR_REDISFAIL, "imToken写入redis失败", nil)
	}

	if "" != oldToken {

		url := "http://106.53.107.155:6666/post_system_message" + "?appid=" + strconv.FormatInt(oneUser.Appid, 10) + "&uid=" + strconv.FormatInt(oneUser.Uid, 10)
		dataMap := make(map[string]KickUser)
		reqJson := KickUser{}
		reqJson.Token = oldToken
		reqJson.Sender = oneUser.Uid
		reqJson.Timestamp = time.Now().Unix()
		dataMap["kick_user"] = reqJson
		req, err := util.Post(dataMap, url)
		if nil != err {
			c.Error(util.ERR_JSON_CONVERT_FAILED, err.Error(), nil)
		}

		res, err := req.Response()
		if nil != err {
			c.Error(util.ERR_SYSTEM, err.Error(), nil)
		}

		if 200 != res.StatusCode {
			c.Error(util.ERR_SYSTEM, "服务器异常", nil)
		}
	}

	MyInfo := MyInfo{}
	MyInfo.Uid = oneUser.Uid
	MyInfo.Token = token
	MyInfo.Nickname = oneUser.Nickname
	MyInfo.HeadImg = oneUser.HeadImg
	MyInfo.Phone = oneUser.Phone
	MyInfo.Sex = oneUser.Sex
	MyInfo.Signature = oneUser.Signature
	MyInfo.RegionId = oneUser.RegionId

	var regionName string

	if oneUser.RegionId != 0 {
		regionName, _ = models.GetRegoinName(oneUser.RegionId)
	} else {
		regionName = ""
	}

	MyInfo.RegionName = regionName

	//内存存用户列表
	UserList[uInfo.Uid] = &MyInfo

	c.Success(MyInfo)
}

/*
	主动验证token接口
*/
// @Title 主动验证token接口
// @Description 主动验证token接口
// @Param	token		header 	string	true "用户token"
// @Success  200  {object}  controllers.MyInfo
// @router /verifyToken [post]
func (c *UserController) VerifyToken() {
	token := c.GetHeaderToken()
	if token == "" {
		c.Error(util.ERR_MISSPARAM, "token参数不能为空", nil)
	}
	//验证token
	_, _, err := AuthToken(token)
	if err != nil {
		c.Error(util.ERR_TOKEN, "token已失效 请重新登录!", nil)
	}
	c.Success(nil)
}

/*
	获取世界地区信息（树形结构）
*/
// @Title 获取世界地区信息
// @Description 获取世界地区信息接口
// @Param	token		header 	string	true "用户token"
// @Success 200 {object} models.[]
// @router /getRegions [post]
func (c *UserController) GetRegions() {
	token := c.GetHeaderToken()
	if token == "" {
		c.Error(util.ERR_MISSPARAM, "token参数不能为空", nil)
	}
	//验证token
	_, _, err := AuthToken(token)
	if err != nil {
		c.Error(util.ERR_TOKEN, "token已失效 请重新登录!", nil)
	}
	c.Success(models.RegionTreeList)
}

//验证token
func AuthToken(token string) (int64, int64, error) {

	appid, uid, err := LoadUserAccessToken(token)

	if err != nil {
		return 0, 0, err
	}

	return appid, uid, nil
}

func LoadUserAccessToken(token string) (int64, int64, error) {
	conn := models.Redis_pool.Get()
	defer conn.Close()

	key := fmt.Sprintf("access_token_%s", token)
	var uid int64
	var appid int64

	err := conn.Send("EXISTS", key)
	if err != nil {
		return 0, 0, err
	}
	err = conn.Send("HMGET", key, "user_id", "app_id")
	if err != nil {
		return 0, 0, err
	}
	err = conn.Flush()
	if err != nil {
		return 0, 0, err
	}

	exists, err := redis.Bool(conn.Receive())
	if err != nil {
		return 0, 0, err
	}
	reply, err := redis.Values(conn.Receive())
	if err != nil {
		return 0, 0, err
	}

	if !exists {
		return 0, 0, errors.New("token non exists")
	}
	_, err = redis.Scan(reply, &uid, &appid)
	if err != nil {
		return 0, 0, err
	}

	return appid, uid, nil
}

//存入redis最新的 IM token
func SetAuthToken(conn redis.Conn, appid int64, user_id int64) (string, string, error) {

	//过期之前redis里面的token
	key1 := fmt.Sprintf("ut_%d_%d", appid, user_id)
	old_token, _ := conn.Do("GET", key1)
	old_token_str := ""

	if old_token != nil {
		old_token_str = string(old_token.([]uint8))
		old_token_key := fmt.Sprintf("access_token_%s", old_token_str)
		conn.Do("PEXPIRE", old_token_key, -1)
	}

	//生成gobelieve所需的登录的token
	token := util.RandomStringWithCharset(24, CHARSET)
	key2 := fmt.Sprintf("access_token_%s", token)
	_, err := conn.Do("HMSET", key2, "access_token", token, "user_id", user_id, "app_id", appid)
	if err != nil {
		logs.Error("HMSET err:", err)
		return "", old_token_str, err
	}
	_, err = conn.Do("PEXPIRE", key2, 1000*3600*24*7)

	if err != nil {
		logs.Error("PEXPIRE err:", err)
		return "", old_token_str, err
	}

	//生成appid下user对应token
	key3 := fmt.Sprintf("ut_%d_%d", appid, user_id)
	conn.Do("SET", key3, token)
	conn.Do("PEXPIRE", key3, 1000*3600*24*7)

	return token, old_token_str, nil

}

/*
	获取所有的用户信息
*/
// @Title 获取所有的用户信息
// @Description 获取所有的用户信息
// @Param	appid		query 	string	true "产品id"
// @Success 200 {string} 123
// @router /listAllUser [get]
func (c *UserController) ListAllUser() {
	appidStr := c.GetString("appid")
	appid, err := strconv.ParseInt(appidStr, 10, 64)
	if nil != err {
		c.Error(util.ERR_TYPECONVERSION, err.Error(), nil)
	}

	users, err := models.GetAllUser(appid)
	if nil != err {
		c.Error(util.ERR_DB, err.Error(), nil)
	}
	c.Success(users)
}


/*
	删除账户接口
*/
// @Title 删除账户接口
// @Description 删除账户接口
// @Param	token		header 	string	true "用户token"
// @Success  200  {string} code：0
// @router /delAccount [post]
func  (c *UserController)DelAccount() {

	token := c.GetHeaderToken()
	if token == "" {
		c.Error(util.ERR_MISSPARAM, "token参数不能为空", nil)
	}
	//验证token
	appid, uid, err := AuthToken(token)
	if err != nil {
		c.Error(util.ERR_TOKEN, "token已失效 请重新登录!", nil)
	}

	//事务处理
	o := orm.NewOrm()
	o.Begin()
	defer func() {
		if e := recover(); e != nil {
			o.Rollback()
			logs.Error(e)
			c.Error(util.ERR_TRANS, "删除账户失败!", nil)
		}
	}()
	//删除user_auth表信息
	_,err = o.Raw("DELETE FROM `user_auth` WHERE `appid` = ? AND `id` = ? ", appid, uid).Exec()
	if err != nil {
		panic(err)
	}
	//删除user_info表信息
	_,err = o.Raw("DELETE FROM `user_info` WHERE `appid` = ? AND `uid` = ? ", appid, uid).Exec()
	if err != nil {
		panic(err)
	}
	//删除好友关系表信息
	_,err = o.Raw("DELETE FROM `friend` WHERE `appid` = ? AND `uid` = ? OR `friend_uid` = ?", appid, uid, uid).Exec()
	if err != nil {
		panic(err)
	}

	//删除组关系表信息
		//1.如果是群组组长 删除所以群信息
	_,err = o.Raw("DELETE FROM `group_member` WHERE `group_id` IN (SELECT `id` FROM `group` WHERE `appid` = ? AND `master` = ?)", appid, uid).Exec()
	if err != nil {
		panic(err)
	}
		//2.删除所以组信息
	_,err = o.Raw("DELETE FROM `group` WHERE `appid` = ? and `master` = ? ", appid, uid).Exec()
	if err != nil {
		panic(err)
	}

	err = nil

	o.Commit()

	c.Success(nil)

}