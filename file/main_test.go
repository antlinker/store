package file_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	. "github.com/antlinker/store"
)

func testUpdate(t *testing.T) {
	data1 := []byte("test123132312123")
	data2 := []byte("test1231323")
	filename := "testUpdate12.txt"
	err := SaveData(filename, data1)
	if err != nil {
		t.Error(err)
		return
	}
	err = SaveData(filename, data2)
	if err != nil {
		t.Error(err)
		return
	}
	err = UpdateData(filename, data2)
	if err != nil {
		t.Error(err)
		return
	}
	e, err := IsExists(filename)
	if err != nil {
		t.Error(err)
		return
	}
	if !e {
		t.Error("文件应该存在")
		return
	}
	size, err := Fsize(filename)
	if err != nil {
		t.Error(err)
		return
	}
	if size != int64(len(data2)) {
		t.Error("获取大小错误")
		return
	}
	err = DeleteFile(filename)
	if err != nil {
		t.Error(err)
		return
	}
	e, err = IsExists(filename)
	if err != nil {

		t.Error(err)
		return
	}
	if e {
		t.Error("文件不应该存在")
		return
	}
}

func testRead(t *testing.T) {
	data1 := []byte("test123132312123")
	filename1 := "test1.txt"
	err := SaveData(filename1, data1)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err := DeleteFile(filename1)
		if err != nil {
			t.Error(err)
		}
	}()

	r, err := DefaultStore.GetReader(filename1)
	if err != nil {
		t.Errorf("失败:%s", err)
		return
	}
	if r == nil {
		t.Errorf("读取失败!")
		return
	}
	defer r.Close()

	buff := bytes.NewBuffer(nil)
	_, err = io.Copy(buff, r)
	if err != nil {
		t.Error(err)
		return
	}
	data := buff.Bytes()

	if bytes.Compare(data1, data) != 0 {
		t.Error("读取的数据不是存入数据:%s", data)
		return
	}

}
func testMultifilePackaging(t *testing.T) {

	data1 := []byte("test123132312123")
	data2 := []byte("test1231323")
	filename1 := "test1.txt"
	filename2 := "test2.txt"

	zipfname := "test.zip"
	err := SaveData(filename1, data1)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err := DeleteFile(filename1)
		if err != nil {
			t.Error(err)
		}
	}()
	err = SaveData(filename2, data2)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err := DeleteFile(filename2)
		if err != nil {
			t.Error(err)
		}
	}()
	tf, err := os.Create(zipfname)
	if err != nil {
		t.Errorf("失败:%s", err)
		return
	}
	err = DefaultStore.MultifilePackaging(tf,
		FileAlias{
			Alias: filename1,
			Key:   filename1,
		},

		FileAlias{
			Alias: filename2,
			Key:   filename2,
		},
	)
	if err != nil {
		t.Errorf("失败:%s", err)
		return
	}
	tf.Close()
	r, err := IsExists(zipfname)
	if err != nil {
		t.Errorf("文件检查失败:%s", err)
		return
	}
	if !r {
		t.Errorf("文件不存在测试失败.")
		return
	}
	err = DeleteFile(zipfname)
	if err != nil {
		t.Error(err)
	}
	r, err = IsExists(zipfname)
	if err != nil {
		t.Errorf("文件检查失败:%s", err)
		return
	}
	if r {
		t.Errorf("文件存在删除失败.")
		return
	}
}
