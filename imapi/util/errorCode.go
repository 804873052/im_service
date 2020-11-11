package util

const (
	//正确返回
	ERR_OK = 0

	//参数错误
	ERR_MISSPARAM = 10001

	//类型转换错误
	ERR_TYPECONVERSION = 10002

	//验证码错误
	ERR_CAPTCHA = 10030

	//登录失败，授权错误
	ERR_TOKEN = 10100

	//无权限操作
	ERR_NOTPERMISSION = 10200

	//系统错误
	ERR_SYSTEM = 10300

	//未查询到用户
	ERR_NOFOUND = 10405

	//数据库错误
	ERR_DB = 10555

	//事务处理错误
	ERR_TRANS = 10500

	//重复注册
	ERR_REPLYREG = 10600

	//用户名不存在
	ERR_NOTFOUND = 11010

	//密码不正确
	ERR_PWDNOTMATCH = 11011

	//添加用户不存在
	ERR_USER_NOTFOUND = 11012

	//redis写入失败
	ERR_REDISFAIL = 12001

	//redis写入失败
	ERR_NOGROUP = 13001

	//请勿重复申请好友
	ERR_REPEAT_APPLY_FRIEND = 13002

	//您不是群主
	ERR_NOT_MASTER = 13003

	//重复添加群成员
	ERR_REPEAT_ADD_GROUP_MEMBER = 13004

	//重复添加群成员
	ERR_JSON_CONVERT_FAILED = 13005
)
