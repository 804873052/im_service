package controllers

import (
	"github.com/astaxie/beego"
)

type BaseController struct {
	beego.Controller
}

type ReturnMsg struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func (this *BaseController) Success(data interface{}) {

	//if data == nil {
	//	data = make(map[string]string)
	//}

	res := ReturnMsg{
		0, "success", data,
	}
	this.Data["json"] = res
	this.ServeJSON() //对json进行序列化输出
	this.StopRun()
}

func (this *BaseController) Error(code int, msg string, data interface{}) {

	res := ReturnMsg{
		code, msg, data,
	}

	this.Data["json"] = res
	this.ServeJSON() //对json进行序列化输出
	this.StopRun()
}

//若获取不到头部的token 返回""; 否则返回对应token值

func (this *BaseController) GetHeaderToken() string{

	return this.Ctx.Request.Header.Get("token")

}