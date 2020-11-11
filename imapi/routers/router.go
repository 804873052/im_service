// @APIVersion 1.0.0
// @Title gobelieve API
// @Description gobelieve 接口API
// @Contact astaxie@gmail.com
package routers

import (
	"imapi/controllers"

	"github.com/astaxie/beego"
)

func init() {
	ns := beego.NewNamespace("/v1",
		beego.NSNamespace("/user",
			beego.NSInclude(
				&controllers.UserController{},
			),
		),
		beego.NSNamespace("/myPage",
			beego.NSInclude(
				&controllers.MyPageController{},
			),
		),
		beego.NSNamespace("/friend",
			beego.NSInclude(
				&controllers.FriendController{},
			),
		),
		beego.NSNamespace("/group",
			beego.NSInclude(
				&controllers.GroupController{},
			),
		),
		beego.NSNamespace("/contactList",
			beego.NSInclude(
				&controllers.ContactListController{},
			),
		),
	)
	beego.AddNamespace(ns)
}
