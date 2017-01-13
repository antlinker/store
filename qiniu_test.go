package store_test

import (
	"fmt"
	"io"
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

func testRead(t *testing.T) {
	r, err := store.DefaultStore.GetReader("test.txt")
	if err != nil {
		t.Errorf("失败:%s", err)
		return
	}
	if r == nil {
		t.Errorf("读取失败!")
		return
	}
	fmt.Println("")
	io.Copy(os.Stdout, r)
	fmt.Println("")
	t.Log("\nok\n")
}

func testMultifilePackaging(t *testing.T) {
	tf, err := os.Create("test.zip")
	if err != nil {
		t.Errorf("失败:%s", err)
		return
	}
	err = store.DefaultStore.MultifilePackaging(tf,
		store.FileAlias{
			Alias: "a/test.txt",
			Key:   "test.txt",
		},

		store.FileAlias{
			Alias: "a/bannerTop.png",
			Key:   "bannerTop.png",
		},
	)
	if err != nil {
		t.Errorf("失败:%s", err)
		return
	}
}
