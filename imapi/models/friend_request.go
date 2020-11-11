package models

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	"strconv"
	"time"
)

const (
	//状态 0表示未处理，1表示同意，2表示拒绝
	FR_STATUS_INIT   = 0
	FR_STATUS_AGREE  = 1
	FR_STATUS_REFUSE = 2
)

type FriendRequest struct {
	Id             int64     `json:"id"`
	Appid          int64     `json:"appid"`
	Uid            int64     `json:"uid"`
	FriendUid      int64     `json:"friend_uid"`
	FriendNickname *string   `json:"friend_nickname"`
	Status         int       `json:"status"`
	CreateTime     time.Time `orm:"-" json:"create_time,-"`
	ModifyTime     time.Time `orm:"-" json:"modify_time,-"`
}

type FriendRequestJson struct {
	Id             int64                  `json:"id"`
	Appid          int64                  `json:"appid"`
	Uid            int64                  `json:"uid"`
	FriendUid      int64                  `json:"friend_uid"`
	HeadImg        string                 `json:"head_img"`
	FriendNickname string                 `json:"friend_nickname"`
	Nickname       string                 `json:"nickname"`
	Status         int64                  `json:"status"`
	CreateTime     string                 `orm:"-" json:"create_time,-"`
	ModifyTime     string                 `orm:"-" json:"modify_time,-"`
	Timestamp      int64                  `json:"timestamp"`
	Msg            []FriendRequestMessage `json:"msg"`
}

type FriendRequestMessage struct {
	Uid        int64  `json:"uid"`
	RequestId  string `json:"request_id"`
	MsgContent string `json:"msg_content"`
}

//指定表名
func (fr *FriendRequest) TableName() string {
	return TableName("friend_request")
}

// 设置引擎为 INNODB
func (fr *FriendRequest) TableEngine() string {
	return "INNODB"
}

func (fr *FriendRequest) GetUserByUid() {
	o := orm.NewOrm()
	o.Read(fr)
}

/**
  查找申请记录
*/
func FindRecordByParam(uid int64, fuid int64, appid int64) error {
	o := orm.NewOrm()
	var friend FriendRequest
	err := o.Raw("SELECT id FROM friend_request WHERE appid = ? and uid = ? and friend_uid = ?", appid, uid, fuid).QueryRow(&friend)
	return err
}

/**
添加好友申请
*/
func InsertRecord(uid int64, fuid int64, appid int64, nickname *string, o orm.Ormer) (int64, error) {
	var frqData FriendRequest
	frqData.Appid = appid
	frqData.Uid = uid
	frqData.FriendUid = fuid
	frqData.FriendNickname = nickname

	id, err := o.Insert(&frqData)
	if err == nil {
		fmt.Println(id)
	}

	return id, err
}

/**
修改好友申请状态
*/
func UpdateRecordMessage(requestId int64, status int, o orm.Ormer) (int64, error) {

	frqData := FriendRequest{Id: requestId}
	if o.Read(&frqData) == nil {
		frqData.Status = status
		if num, err := o.Update(&frqData); err == nil {
			return num, err
		}
	}

	return 0, nil
}

/**
  我的新的朋友列表
*/
func GetMyNewFriendsList(uid int64) []FriendRequestJson {
	// todo
	var maps []orm.Params
	o := orm.NewOrm()
	num, _ := o.Raw("SELECT fr.*, u.head_img, u.nickname FROM friend_request fr LEFT JOIN user_info u ON fr.uid = u.uid WHERE fr.friend_uid = ?", uid).Values(&maps)

	type friendRequestMessage []FriendRequestMessage
	type friendRequestJsonData []FriendRequestJson
	fmt.Println(num)
	if num <= 0 {
		return []FriendRequestJson{}
	}

	var requestId = []int64{}
	friendRequestJsonSlice := friendRequestJsonData{}
	friendRequestMessageSlice := friendRequestMessage{}
	for _, v := range maps {
		id := v["id"]
		appid := v["appid"]
		uid := v["uid"]
		fuid := v["friend_uid"]
		headImg := v["head_img"]
		friendNickname := v["friend_nickname"]
		nickname := v["nickname"]
		status := v["status"]
		createTime := v["create_time"]
		modifyTime := v["modify_time"]
		ids, _ := strconv.ParseInt(id.(string), 10, 64)
		appids, _ := strconv.ParseInt(appid.(string), 10, 64)
		uids, _ := strconv.ParseInt(uid.(string), 10, 64)
		fuids, _ := strconv.ParseInt(fuid.(string), 10, 64)
		statuss, _ := strconv.ParseInt(status.(string), 10, 64)
		headImgs := headImg.(string)
		friendNicknames := friendNickname.(string)
		nicknames := nickname.(string)
		createTimes := createTime.(string)
		modifyTimes := modifyTime.(string)

		//requestId = append(requestId, ids.(int64))

		friendRequestJsonSlice = append(friendRequestJsonSlice, FriendRequestJson{
			Id:             ids,
			Appid:          appids,
			Uid:            uids,
			FriendUid:      fuids,
			HeadImg:        headImgs,
			FriendNickname: friendNicknames,
			Nickname:       nicknames,
			Status:         statuss,
			CreateTime:     createTimes,
			ModifyTime:     modifyTimes,
		})
	}
	for _, v := range maps {
		id := v["id"]

		id, _ = strconv.ParseInt(id.(string), 10, 64)
		requestId = append(requestId, id.(int64))
	}
	if len(requestId) > 0 {
		reqMessage := []FriendMessage{}
		qs2 := o.QueryTable("request_message")
		num, _ = qs2.Filter("request_id__in", requestId).All(&reqMessage)
		if num > 0 {
			timeTemplate1 := "2006-01-02 15:04:05" //常规类型
			for k, v := range friendRequestJsonSlice {
				stamp, _ := time.ParseInLocation(timeTemplate1, v.CreateTime, time.Local)
				friendRequestJsonSlice[k].Timestamp = stamp.Unix()
				for _, vv := range reqMessage {
					if vv.RequestId == v.Id {
						friendRequestJsonSlice[k].Msg = append(friendRequestMessageSlice, FriendRequestMessage{
							Uid:        vv.Uid,
							RequestId:  strconv.FormatInt(vv.RequestId, 10),
							MsgContent: vv.MsgContent,
						})
					}

				}
			}
		}
	}
	return friendRequestJsonSlice
}


/**
  我的新的朋友列表
*/
func GetMyNewFriendsListV2(uid int64) (friendRequestJson []FriendRequestJson,err error)  {
	//var maps []orm.Params
	o := orm.NewOrm()
	sql := `SELECT a.*, b.head_img, b.nickname FROM friend_request a
			INNER JOIN (
				SELECT
					max(fr.id) id,
					fr.uid,
					u.head_img,
					u.nickname
				FROM
					friend_request fr
				LEFT JOIN user_info u ON fr.uid = u.uid
				WHERE
					fr.friend_uid = ?
				GROUP BY
				fr.uid
			) b ON a.uid = b.uid
			AND a.id = b.id`
	num, err := o.Raw(sql, uid).QueryRows(&friendRequestJson)
	//fmt.Println(num)
	if num == 0 {
		return nil,nil
	}
	return
}

/**
  新的朋友消息
*/
func GetMyNewFriendsMsg(rid int64) []FriendRequestMessage {
	var friendMsgs []FriendRequestMessage
	o := orm.NewOrm()
	_, err := o.Raw("SELECT request_id, uid, msg_content FROM request_message WHERE request_id = ?", rid).QueryRows(&friendMsgs)
	if err != nil {
	}
	return friendMsgs
}
