package qiniu

import (
	"fmt"
	"log"

	"qiniupkg.com/api.v7/conf"
	"qiniupkg.com/api.v7/kodo"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	dburl     string
	dbname    string
	poolLimit int
	session   *mgo.Session
	collname  = "qiniucfg"
)

// MgoInitConfig 配置mgodb
func mgoInitConfig(url, name string) {
	dburl = url
	dbname = name
	poolLimit = 2
	log.Printf("MGO_URL = %s", dburl)
	log.Printf("MGO_DB_NAME = %s", dbname)
}

// GetSession 获取Mongodb的session，如果为空则重新创建，否则返回原有值
func getSession() (*mgo.Session, error) {
	if session == nil {
		var err error
		session, err = mgo.Dial(dburl)
		if err != nil {
			log.Printf("连接mongodb错误:%s", err.Error())
			return nil, err
		}
		session.SetPoolLimit(poolLimit)
	}
	return session.Clone(), nil
}

// CreateKeyManagerByMGO 创建七牛云密钥管理
func CreateKeyManagerByMGO(url, name string) KeyManager {
	mgoInitConfig(url, name)
	return &qiniuKeyManager{
		syncer:  &qiniuKeySyncByMgo{},
		updater: &qiniuKeyUpdateByMgo{},
	}
}

// StartKeyManagerByMGO 开始同步七牛云密钥
func StartKeyManagerByMGO(url, name string) {
	mgoInitConfig(url, name)
	defaultKeyManager = &qiniuKeyManager{
		syncer:  &qiniuKeySyncByMgo{},
		updater: &qiniuKeyUpdateByMgo{},
	}
	defaultKeyManager.StartSync()
}

type qiniuKeySyncByMgo struct {
	AK      string `bson:"ak"`
	SK      string `bson:"sk"`
	Version int    `bson:"version"`
	curv    int
}

func (m *qiniuKeySyncByMgo) Sync() {
	sess, err := getSession()
	if err != nil {
		fmt.Println("连接mgodb数据库失败:", err)
		return
	}
	coll := sess.DB(dbname).C(collname)
	err = coll.Find(bson.M{"_id": 1}).One(m)
	if err != nil {
		fmt.Println("查询密钥失败：", err)
		return
	}
	fmt.Println("查询密钥：", m)
	if m.curv != m.Version {
		conf.ACCESS_KEY = m.AK
		conf.SECRET_KEY = m.SK
		kodo.SetMac(m.AK, m.SK)
		m.curv = m.Version
	}
}

type qiniuKeyUpdateByMgo struct {
	qiniuKeySyncByMgo
}

func (m *qiniuKeyUpdateByMgo) Update(ak, sk string) error {
	sess, err := getSession()
	if err != nil {
		log.Print("连接mgodb数据库失败:", err)
		return err
	}
	coll := sess.DB(dbname).C(collname)
	_, err = coll.UpsertId(1, bson.M{"ak": ak, "sk": sk, "$inc": bson.M{"version": 1}})

	if err != nil {
		log.Print("存储密钥失败：", err)
		return err
	}
	return nil
}
