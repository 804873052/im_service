package models

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"net/url"
	"strconv"
	"time"
)

var Redis_pool *redis.Pool

func init() {
	dbhost := beego.AppConfig.String("db.host")
	dbport := beego.AppConfig.String("db.port")
	dbuser := beego.AppConfig.String("db.user")
	dbpassword := beego.AppConfig.String("db.password")
	dbname := beego.AppConfig.String("db.dbname")
	timezone := beego.AppConfig.String("db.timezone")
	if dbport == "" {
		dbport = "3306"
	}

	dsn := dbuser + ":" + dbpassword + "@tcp(" + dbhost + ":" + dbport + ")/" + dbname + "?charset=utf8"

	if timezone != "" {
		dsn = dsn + "&loc=" + url.QueryEscape(timezone)
	}
	//dsn = dsn + "&parseTime=true"
	//fmt.Println(dsn)
	//dsn := "root:123456@tcp(192.168.131.144)/gobelieve?charset=utf8"

	orm.RegisterDataBase("default", "mysql", dsn, 30)

	orm.RegisterModel(
		new(UserAuth),
		new(UserInfo),
		new(Region),
		new(Feedback),
		new(FriendRequest),
		new(FriendMessage),
		new(Friend),
		new(Group),
		new(GroupMember))

	if beego.AppConfig.String("runmode") == "dev" {
		orm.Debug = true
	}

	//orm.RunSyncdb("default",false,true)
	//redis 连接池初始化
	conn := beego.AppConfig.String("cache.conn")
	dbNumStr := beego.AppConfig.String("cache.dbNum")
	dbNum, _ := strconv.Atoi(dbNumStr)
	password := beego.AppConfig.String("cache.password")
	Redis_pool = NewRedisPool(conn, password, dbNum)

	//初始化region世界地区信息
	go InitTreeList()
}

func TableName(name string) string {
	return beego.AppConfig.String("db.prefix") + name
}

func NewRedisPool(server, password string, db int) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     100,
		MaxActive:   500,
		IdleTimeout: 480 * time.Second,
		Dial: func() (redis.Conn, error) {
			timeout := time.Duration(2) * time.Second
			c, err := redis.DialTimeout("tcp", server, timeout, 0, 0)
			if err != nil {
				return nil, err
			}
			if len(password) > 0 {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			if db > 0 && db < 16 {
				if _, err := c.Do("SELECT", db); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
	}
}
