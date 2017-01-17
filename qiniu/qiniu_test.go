package qiniu_test

import (
	"os"
	"testing"

	"github.com/antlinker/store"
)

func TestQiniuUpdate(t *testing.T) {
	//
	// StartQiniuKeyManagerByMGO("", "")
	// CreDefaultQiniuStore(QiniuMgoCfg{
	// 	MgoURL: "",
	// 	Dbname: "ant",
	// 	Domain: "odufa0grd.bkt.clouddn.com",
	// 	Bucket: "static-test",
	// })
	store.InitQiniuStore("", "", "mytest")
	store.SetVisitHTTPBase("")
	//	testUpdate(t)
	//	testReaderUpdate(t)
	//getToken(t)
	testRead(t)
	testMultifilePackaging(t)
}

func getToken(t *testing.T) {
	t.Log(store.GetVisitURL("1464256-6373bdb2116a78c5.png-imageView2/0/w/48/h/48"))
}
func testReaderUpdate(t *testing.T) {
	filename := "README.md"
	dstfile := "README2.md"
	f, _ := os.Open(filename)
	s, _ := os.Stat(filename)
	err := store.SaveReader(filename, f, s.Size())
	if err != nil {
		t.Error(err)
		return
	}

	e, err := store.IsExists(filename)
	if err != nil {
		t.Error(err)
		return
	}
	if !e {
		t.Error("文件应该存在")
		return
	}
	size, err := store.Fsize(filename)
	if err != nil {
		t.Error(err)
		return
	}
	if size != int64(s.Size()) {
		t.Error("获取大小错误")
		return
	}
	store.Move(filename, dstfile)
}
