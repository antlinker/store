package file

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/antlinker/store"
)

// InitStore 初始化文件存储
func InitStore(rootDir string) error {
	ds, err := CreateStore(rootDir)
	store.DefaultStore = ds
	return err
}

// CreateStore 创建一个文件存储
func CreateStore(rootDir string) (store.Storer, error) {
	r, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("设置根目录失败:%s", err)
	}
	return &fileStore{rootDir: r}, nil
}

type fileStore struct {
	rootDir string
	base    string
}

func mkdirs(filename string) error {
	return os.MkdirAll(path.Dir(filename), os.ModePerm)
}
func (s *fileStore) abs(dst string) string {

	n := "/" + dst
	fname := filepath.Base(n)
	path := filepath.Dir(n)
	if path == "/" && fname == "/" {
		return s.rootDir
	}
	//fmt.Println("abs dst:", s.rootDir,dst)

	if path == "/" {
		//fmt.Println("abs:", s.rootDir + path  + fname)
		return s.rootDir + path + fname
	}
	fmt.Println("abs:", s.rootDir+path+"/"+fname)
	return s.rootDir + path + "/" + fname

}
func (s *fileStore) SetVisitHTTPBase(path string) {
	s.base = path
}
func (s *fileStore) GetVisitURL(key string) string {
	return fmt.Sprintf("%s/%s", s.base, key)
}
func (s *fileStore) Copy(src, dst string) error {
	src = s.abs(src)

	dst = s.abs(dst)

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)

	if err != nil {
		return err
	}

	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
func (s *fileStore) Move(src, dst string) error {
	src = s.abs(src)

	dst = s.abs(dst)
	return os.Rename(src, dst)
}

// SaveFile 将数据data存储到filenamet
func (s *fileStore) SaveData(filename string, data []byte) (err error) {
	filename = s.abs(filename)

	mkdirs(filename)
	return ioutil.WriteFile(filename, data, os.ModePerm)
}
func (s *fileStore) SaveReader(filename string, data io.Reader, size int64) (err error) {
	filename = s.abs(filename)
	mkdirs(filename)
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	buff := make([]byte, 1024)
	sum := int(size)
	for {
		n, e1 := data.Read(buff)
		if e1 != nil {
			err = e1
			break
		}
		if n >= sum {
			wn, e := f.Write(buff[0:sum])
			if e != nil {
				err = e
				break
			}
			if wn < sum {
				err = io.ErrShortWrite
				break
			}
			break
		}
		wn, e := f.Write(buff[0:n])
		if e != nil {
			err = e
			break
		}
		if wn < n {
			err = io.ErrShortWrite
			break
		}
		sum -= n

	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}
func (s *fileStore) SaveFile(filename string, srcfile string) (err error) {
	filename = s.abs(filename)
	mkdirs(filename)
	f, err := os.OpenFile(srcfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	inf, err := f.Stat()
	if err != nil {
		return err
	}
	err = s.SaveReader(filename, f, inf.Size())
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}

// DeleteFile 删除指定文件
func (s *fileStore) DeleteFile(filename string) (err error) {
	return os.Remove(filename)
}

func (s *fileStore) UpdateData(filename string, data []byte) error {
	return s.SaveData(filename, data)
}
func (s *fileStore) UpdateFile(filename string, srcfile string) (err error) {
	return s.SaveFile(filename, srcfile)
}
func (s *fileStore) UpdateReader(filename string, data io.Reader, size int64) (err error) {
	return s.SaveReader(filename, data, size)
}
func (s *fileStore) IsExists(filename string) (bool, error) {
	filename = s.abs(filename)
	_, err := os.Stat(filename)
	if err != nil {
		r := os.IsNotExist(err)
		if r {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
func (s *fileStore) Fsize(filename string) (int64, error) {
	filename = s.abs(filename)
	inf, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	return inf.Size(), nil

}
func (s *fileStore) PutTime(filename string) (time.Time, error) {
	filename = s.abs(filename)
	inf, err := os.Stat(filename)
	if err != nil {
		return time.Time{}, err
	}
	return inf.ModTime(), nil
}

// 获取文件流
// 从本地获取一个可以关闭的文件流
func (s *fileStore) GetReader(key string) (io.ReadCloser, error) {
	key = s.abs(key)
	return os.Open(key)
}

// 文件打包
// packfile 返回打包文件路径
func (s *fileStore) MultifilePackaging(w io.Writer, keys ...store.FileAlias) (err error) {
	//buffer := new(bytes.Buffer)
	writer := zip.NewWriter(w)
	defer writer.Close()

	var errInfo error
	for _, file := range keys {
		w, err := writer.CreateHeader(&zip.FileHeader{
			Name:   file.Alias,
			Flags:  1 << 11,
			Method: zip.Deflate,
		})
		if err != nil {
			errInfo = err
			break
		}

		f, err := os.Open(s.abs(file.Key))
		if err != nil {
			errInfo = err
			f.Close()
			break
		}
		_, err = io.Copy(w, f)
		if err != nil {
			errInfo = err
			f.Close()
			break
		}
		f.Close()
	}
	return errInfo
}

func (s *fileStore) GetImageInfo(key string) (ii *store.ImageInfo, err error) {
	err = errors.New("文件存储暂不支持")
	return
}

func (s *fileStore) SaveReaderAt(filename string, data io.ReaderAt, size int64) (err error) {
	err = errors.New("文件存储暂不支持")
	return
}
