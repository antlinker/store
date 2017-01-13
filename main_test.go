package store_test

import (
	"os"
	"testing"

	. "github.com/antlinker/store"
)

func testUpdate(t *testing.T) {
	data1 := []byte("test123132312123")
	data2 := []byte("test1231323")
	filename := "test12.txt"
	err := SaveData(filename, data1)
	if err != nil {
		t.Error(err)
	}
	err = SaveData(filename, data2)
	if err != nil {
		t.Error(err)
	}
	err = UpdateData(filename, data2)
	if err != nil {
		t.Error(err)
	}
	e, err := IsExists(filename)
	if err != nil {
		t.Error(err)
	}
	if !e {
		t.Error("文件应该存在")
	}
	size, err := Fsize(filename)
	if err != nil {
		t.Error(err)
	}
	if size != int64(len(data2)) {
		t.Error("获取大小错误")
	}
	err = DeleteFile(filename)
	if err != nil {
		t.Error(err)
	}
	e, err = IsExists(filename)
	if err != nil {

		t.Error(err)
	}
	if e {
		t.Error("文件不应该存在")
	}
}
func testReaderUpdate(t *testing.T) {
	filename := "README.md"
	dstfile := "README2.md"
	f, _ := os.Open(filename)
	s, _ := os.Stat(filename)
	err := SaveReader(filename, f, s.Size())
	if err != nil {
		t.Error(err)
	}

	e, err := IsExists(filename)
	if err != nil {
		t.Error(err)
	}
	if !e {
		t.Error("文件应该存在")
	}
	size, err := Fsize(filename)
	if err != nil {
		t.Error(err)
	}
	if size != int64(s.Size()) {
		t.Error("获取大小错误")
	}
	Move(filename, dstfile)
}
