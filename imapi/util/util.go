package util

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"strconv"
	"time"
	"unsafe"
)


/*
 * 返回随机字符串
 */

func GetToken(uid int) string {
	return  fmt.Sprintf("%d%s",time.Now().Unix(),RandSeq(5))
}
/*
 * 时间戳转日期函数
 */
func UnixToDate(timestamp int) string {
	t := time.Unix(int64(timestamp),0)
	return t.Format("2006-01-02 15:04:05")
}

/*
 * 日期转时间戳函数
 */
func DateToUnix(str string) int64 {
	template := "2006-01-02 15:04:05"
	t,err := time.ParseInLocation(template,str,time.Local)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	return t.Unix()
}

/*
 * 格式化日期 str
 */
func FormatDate(str string) string {

	template := "2006-01-02 15:04:05"
	t,err := time.ParseInLocation(template,str,time.Local)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	if t.IsZero() {
		return ""
	}
	return t.Format(template)

}

/*
 * 获取当前时间戳
 */
func GetUnix() int64 {
	return time.Now().Unix()
}

/*
 * 获取当前日期
 */
func GetDate() string {
	template := "2006-01-02 15:04:05"
	return time.Now().Format(template)
}

/*
 * string 转换 int 类型数据
 */
func StrToInt(str string) int {
	value_int64,_ :=strconv.ParseInt(str,10,64)
	value_int := *(*int)(unsafe.Pointer(&value_int64))
	return value_int
}

/*
 * MD5加密
 */
func MD5Encode(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has) //将[]byte转成16进制
	return md5str
}


/*
 * 返回随机字符串
 */

func RandSeq(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

var seededRand *rand.Rand
/*
 * 返回随机字符串
 */
func RandomStringWithCharset(length int, charset string) string {
	seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}