package util

import (
	xml2 "encoding/xml"
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

/**
 * 发送短信验证码
 * @param string phone
 * @param string msg
 * @return error
 */
const (
	SMSURL =  "http://smsapi.5taogame.com/sms/httpSmsInterface?action=sendhy"
	SMSUERID = "jiyouwang"
	SMSACCOUNT = "jiyouwang"
	SMSPASSWORD = "jiyouwang"
	SMSTEMPLATE = "【机友科技】 验证码为："
)

//短信平台返回的xml结构体
type Resp struct {
	XMLName  xml2.Name `xml:"returnsms"` // 指定最外层的标签为config
	Returnstatus string `xml:"returnstatus"`
	Message string `xml:"message"`
	Remainpoint string `xml:"remainpoint"`
	TaskID string `xml:"taskID"`
	SuccessCounts string `xml:"successCounts"`
}

func SendSmsCaptcha(phone string) (string,error) {

	logs.SetLogger(logs.AdapterFile, `{"filename":"sms.log"}`)

	var content string

	Captcha := CreateCaptcha(6)
	content = fmt.Sprintf(SMSTEMPLATE + "%s",Captcha)
	//存redis todo

	//UrlEncode
	urlEncodeContent := url.QueryEscape(content)
	URL := fmt.Sprintf(SMSURL+"&userid=%s&account=%s&password=%s&mobile=%s&content=%s",SMSUERID,SMSACCOUNT,SMSPASSWORD,phone,urlEncodeContent)
	logs.Info(URL)

	//发送http请求（get）
	resp,err := HttpGet(URL)
	//解析xml的字符串数据
	v := Resp{}
	err = xml2.Unmarshal([]byte(resp), &v)
	if err != nil {
		logs.Error(err)
		return "",err
	}
	logs.Info(v)
	if v.Returnstatus != "Success" {
		return "",errors.New("短信平台处理失败！")
	}
	return Captcha,nil
}

//生成随机位数的短信验证码
func CreateCaptcha(num int) string {
	format := "%0"+  strconv.Itoa(num) + "v"
	format_num := math.Pow(10,float64(num))
	return fmt.Sprintf(format, rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(int32(format_num)))
}

//HTTP get请求
func HttpGet(url string) (string, error) {

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		return "",err
	}
	if resp == nil {
		return "",nil
	}
	defer resp.Body.Close()
	r, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(r), nil

}