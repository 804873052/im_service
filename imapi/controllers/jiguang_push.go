package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"imapi/models"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/toolbox"
	"github.com/gomodule/redigo/redis"
	"github.com/ylywyn/jpush-api-go-client"
	"time"
)

type JPushController struct {
	BaseController
}

const (
	appKey = "c830b26460aa5bceb53654a4"
	secret = "dc030521923f30c2c88f8aaa"
	QUEUENAME = "push_queue"
	POPBACTHNUM = 10
)

/*
	定时任务测试（定时读取redis离线消息数据）
*/

func PopQueue() error {

	getOfflineMsgs,err := getOfflineMsgs()
	if err != nil {
		return err
	}
	//logs.Info(getOfflineMsgs)
	getUserInfo(getOfflineMsgs)
	return nil
}
//todo
func getUserInfo(offlineMsgs [][]byte)  {
	type test struct {
		Appid int
		Content string
		Receiver int
		Sender int
	}
	var class test
	for _,v := range offlineMsgs {
		err := json.Unmarshal(v,&class)
		fmt.Println(err)
		fmt.Println(class)
	}

}

func getOfflineMsgs() ([][]byte,error) {

	//redis中获取队列离线消息
	conn := models.Redis_pool.Get()
	defer conn.Close()

	num,err := conn.Do("LLEN", QUEUENAME)
	if err != nil{
		logs.Error("redis get push_queue len err: ",err.Error())
		return nil,err
	}else{
		logs.Info("push_queue's len is ",num)
		if num.(int64) == 0 {
			return nil,errors.New("无数据需要处理")
		}
	}
	//一次性处理离线消息的数量
	var popBatchNum int64

	if num.(int64) > POPBACTHNUM {
		popBatchNum = POPBACTHNUM
	}else{
		popBatchNum = num.(int64)
	}

	logs.Info("popBatchNum: ",popBatchNum)
	//需要处理的离线消息切片
	offlineMsgs := make([][]byte,0)

	begin := time.Now()

	for i := 0;i < int(popBatchNum); i++ {
		reply,err :=  redis.Bytes(conn.Do("LPOP", QUEUENAME))
		if err != nil {
			logs.Error("lpop error:", err)
			return nil,err
		}
		offlineMsgs = append(offlineMsgs,reply)
	}

	//logs.Info("offlineMsgs:", offlineMsgs)
	end := time.Now()
	duration := end.Sub(begin)
	logs.Info("lpop:%d time:%s success", popBatchNum, duration)

	//if  duration > time.Millisecond*300 {
	//	logs.Error("multi lpop slow:", duration)
	//}

	return offlineMsgs,nil
}

func Jcrond() {
	tk := toolbox.NewTask("task1", "0/5 * * * * *", PopQueue)
	err := tk.Run()
	if err != nil{
		fmt.Println(err)
	}
	toolbox.AddTask("task1", tk)
	toolbox.StartTask()
	//time.Sleep(time.Minute * 5)
	//toolbox.StopTask()

}

/*
	极光推送测试
*/

func (J *JPushController)Jtest() {

	//Platform
	var pf jpushclient.Platform
	pf.Add(jpushclient.ANDROID)
	pf.Add(jpushclient.IOS)
	//pf.Add(jpushclient.WINPHONE)
	//pf.All()

	//Audience
	var ad jpushclient.Audience
	//s := []string{"1", "2", "3"}
	s := []string{"480214e48d17b021b3508eb17f1f1d0a"}

	//ad.SetTag(s)
	ad.SetAlias(s)
	//ad.SetID(s)
	//ad.All()

	//Notice
	var notice jpushclient.Notice
	notice.SetAlert("alert_test123")
	extra := make(map[string]interface{})
	extra["kevin"] = "kevin1"
	extra["test"] = "haha"

	notice.SetAndroidNotice(&jpushclient.AndroidNotice{Alert: "您有新的订单，请及时查看~",BuilderId: 2,Extras:extra})
	notice.SetIOSNotice(&jpushclient.IOSNotice{Alert: "IOSNotice123"})
	//notice.SetWinPhoneNotice(&jpushclient.WinPhoneNotice{Alert: "WinPhoneNotice"})

	var msg jpushclient.Message
	msg.Title = "Hello1"
	msg.Content = "你是ylywn2"

	payload := jpushclient.NewPushPayLoad()
	payload.SetPlatform(&pf)
	payload.SetAudience(&ad)
	payload.SetMessage(&msg)
	payload.SetNotice(&notice)
	//option
	var option jpushclient.Option
	option.SetApns(true)
	payload.SetOptions(&option)

	bytes, _ := payload.ToBytes()
	fmt.Printf("%s\r\n", string(bytes))

	//push
	c := jpushclient.NewPushClient(secret, appKey)
	str, err := c.Send(bytes)
	if err != nil {
		fmt.Printf("err:%s", err.Error())
	} else {
		fmt.Printf("ok:%s", str)
	}
}