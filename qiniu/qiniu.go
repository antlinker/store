package qiniu

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/antlinker/store"

	"regexp"

	"qiniupkg.com/api.v7/conf"
	"qiniupkg.com/api.v7/kodo"
	"qiniupkg.com/api.v7/kodocli"
	reqid "qiniupkg.com/x/reqid.v7"
	"qiniupkg.com/x/rpc.v7"
	"qiniupkg.com/x/url.v7"
)

// MgoCfg 七牛云配置
type MgoCfg struct {
	MgoURL         string
	Dbname         string
	Domain         string
	Bucket         string
	ImageInfoStyle string // 图片信息获取样式名称
}

// CreDefaultStoreByMGO 创建七牛云存储支持
func CreDefaultStoreByMGO(cfg MgoCfg) {

	StartKeyManagerByMGO(cfg.MgoURL, cfg.Dbname)
	store.DefaultStore = CreateStore(cfg.Bucket, 3600)
	store.DefaultStore.SetVisitHTTPBase(cfg.Domain)
}

var (
	defaultKeyManager KeyManager
	imageInfoStyle    string
)

// UpdateKey 存储新的密钥
// 需要现创建密钥管理器
func UpdateKey(ak, sk string) {
	if defaultKeyManager != nil {
		defaultKeyManager.Update(ak, sk)
	} else {
		log.Printf("未创建密钥管理器")
	}
}

// KeyManager 七牛云密钥管理
type KeyManager interface {
	KeyUpdater
	StartSync()
}

// KeySyncer 七牛云密钥同步
type KeySyncer interface {
	Sync()
}

// KeyUpdater 七牛云密钥存储更新
type KeyUpdater interface {
	Update(ak, sk string) error
}

type qiniuKeyManager struct {
	syncer  KeySyncer
	updater KeyUpdater
}

func (m *qiniuKeyManager) Update(ak, sk string) error {
	return m.updater.Update(ak, sk)
}
func (m *qiniuKeyManager) StartSync() {

	m._sync()

}
func (m *qiniuKeyManager) _sync() {
	m.syncer.Sync()
	time.AfterFunc(time.Minute, m._sync)
}

// InitStore 初始化为七牛存储
func InitStore(ak, ck, bucket string) {
	conf.ACCESS_KEY = ak
	conf.SECRET_KEY = ck
	kodo.SetMac(ak, ck)
	store.DefaultStore = CreateStore(bucket, 3600)
}

// CreateStore 创建七牛存储
func CreateStore(bucket string, expires int) store.Storer {

	s := &qiniuStore{
		bucket:  bucket,
		cli:     kodo.New(0, nil),
		expires: expires,
	}
	return s
}

type qiniuStore struct {
	bucket  string
	cli     *kodo.Client
	expires int
	base    string
}

func (s *qiniuStore) SetVisitHTTPBase(path string) {
	if ok, err := regexp.MatchString("^https?://", path); err == nil && ok {
		s.base = path
	} else {
		s.base = "http://" + path
	}
}
func (s *qiniuStore) GetVisitURL(key string) string {
	baseURL := s.base + "/" + url.Escape(key)
	policy := kodo.GetPolicy{}
	//生成一个client对象
	c := kodo.New(0, nil)
	//调用MakePrivateUrl方法返回url
	return c.MakePrivateUrl(baseURL, &policy)
}
func (s *qiniuStore) Stat(key string) (kodo.Entry, error) {
	c := kodo.New(0, nil)
	p := c.Bucket(s.bucket)
	//调用Stat方法获取文件的信息
	return p.Stat(nil, key)

}

func (s *qiniuStore) PutTime(key string) (time.Time, error) {
	e, err := s.Stat(key)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, e.PutTime*100), nil
}

func (s *qiniuStore) Copy(src, dst string) error {
	//new一个Bucket管理对象
	c := kodo.New(0, nil)
	p := c.Bucket(s.bucket)

	//调用Copy方法移动文件
	res := p.Copy(nil, src, dst)

	//打印返回值以及出错信息
	if res == nil {
		return nil
	}
	return fmt.Errorf("Copy failed:%s", res)

}
func (s *qiniuStore) Move(src, dst string) error {
	//new一个Bucket管理对象
	c := kodo.New(0, nil)
	p := c.Bucket(s.bucket)

	//调用Copy方法移动文件
	res := p.Move(nil, src, dst)

	//打印返回值以及出错信息
	if res == nil {
		return nil
	}
	return fmt.Errorf("Copy failed:%s", res)

}

func (s *qiniuStore) SaveFile(filename string, srcfile string) (err error) {
	token := s.getUploadToken()
	return s._saveFile(filename, srcfile, token)
}
func (s *qiniuStore) UpdateFile(filename string, srcfile string) (err error) {
	token := s.getReplaceToken(filename)
	return s._saveFile(filename, srcfile, token)
}
func (s *qiniuStore) _saveFile(filename string, srcfile string, token string) (err error) {
	zone := 0
	uploader := kodocli.NewUploader(zone, nil)

	// if size > chunkSize {
	// 	return uploader.Rput(nil, nil, token, filename, data, size, rputExtra)
	// }
	var ret kodocli.PutRet
	f, err := os.Stat(srcfile)
	if err != nil {
		return err
	}
	if f.Size() > chunkSize {
		//需要分块上传
		uploader.RputFile(nil, &ret, token, filename, srcfile, &rputExtra)
		return
	}

	err = uploader.PutFile(nil, &ret, token, filename, srcfile, nil)
	//打印出错信息
	if err != nil {
		if errCodeMatch(err, 614) {
			s._saveFile(filename, srcfile, s.getReplaceToken(filename))
		}
	}
	return
}

func (s *qiniuStore) SaveData(filename string, data []byte) (err error) {
	token := s.getUploadToken()
	if token == "" {
		return errors.New("获取token失败")
	}
	return s._saveData(filename, data, token)
}
func (s *qiniuStore) UpdateData(filename string, data []byte) (err error) {
	token := s.getReplaceToken(filename)
	if token == "" {
		return errors.New("获取token失败")
	}
	return s._saveData(filename, data, token)
}

func (s *qiniuStore) SaveReader(filename string, data io.Reader, size int64) (err error) {
	token := s.getUploadToken()
	return s._saveReader(filename, data, size, token)
}
func (s *qiniuStore) UpdateReader(filename string, data io.Reader, size int64) (err error) {
	token := s.getReplaceToken(filename)

	return s._saveReader(filename, data, size, token)
}

// 获取临时文件
// 如果是文件系统存储时可能是自身
// 如果是云存储时是一个本缓存文件,如果缓存不存在会创建缓存
func (s *qiniuStore) GetReader(key string) (io.ReadCloser, error) {
	return s.getReader(key)
}
func (s *qiniuStore) getReader(key string) (io.ReadCloser, error) {
	url := s.GetVisitURL(key)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == 200 {

		return res.Body, nil
	}
	return nil, fmt.Errorf("获取远程文件(%s)失败:%d", url, res.StatusCode)
}

// 文件打包
// packfile 返回打包文件路径
func (s *qiniuStore) MultifilePackaging(w io.Writer, keys ...store.FileAlias) (err error) {
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
		f, err := s.getReader(file.Key)
		if err != nil {
			errInfo = err
			break
		}
		io.Copy(w, f)
		f.Close()
	}
	return errInfo
}
func (s *qiniuStore) _saveData(filename string, data []byte, token string) (err error) {

	buff := bytes.NewBuffer(data)
	size := int64(len(data))
	err = s._saveReader(filename, buff, size, token)
	//打印出错信息
	if err != nil {
		if errCodeMatch(err, 614) {
			return s._saveData(filename, data, s.getReplaceToken(filename))
		}
	}
	return
}

const (
	chunkSize int64 = 1024 * 1024 * 4
)

// type RputExtra struct {
// 	Params     map[string]string                             // 可选。用户自定义参数，以"x:"开头 否则忽略
// 	MimeType   string                                        // 可选。
// 	ChunkSize  int                                           // 可选。每次上传的Chunk大小
// 	TryTimes   int                                           // 可选。尝试次数
// 	Progresses []BlkputRet                                   // 可选。上传进度
// 	Notify     func(blkIdx int, blkSize int, ret *BlkputRet) // 可选。进度提示（注意多个block是并行传输的）
// 	NotifyErr  func(blkIdx int, blkSize int, err error)
// }
var (
	rputExtra = kodocli.RputExtra{
		ChunkSize: int(chunkSize),
		TryTimes:  10,
		NotifyErr: func(blkIdx int, blkSize int, err error) {

		},
	}
)

func (s *qiniuStore) _saveReader(filename string, data io.Reader, size int64, token string) (err error) {
	zone := 0
	uploader := kodocli.NewUploader(zone, nil)
	var ret kodocli.PutRet
	return uploader.Put(nil, &ret, token, filename, data, size, nil)

}
func (s *qiniuStore) IsExists(filename string) (bool, error) {
	p := s.cli.Bucket(s.bucket)
	_, err := p.Stat(nil, filename)
	//打印出错时返回的信息
	if err != nil {
		if errCodeMatch(err, 612) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
func (s *qiniuStore) Fsize(filename string) (int64, error) {

	entry, err := s.Stat(filename)
	//打印出错时返回的信息
	if err != nil {
		return 0, err
	}
	fmt.Println(entry)
	return entry.Fsize, nil
}
func (s *qiniuStore) DeleteFile(filename string) (err error) {
	p := s.cli.Bucket(s.bucket)
	//调用Delete方法删除文件
	return p.Delete(nil, filename)

}

func (s *qiniuStore) getUploadToken() (token string) {
	return s.cli.MakeUptoken(&kodo.PutPolicy{
		Scope: s.bucket,
		//设置Token过期时间
		Expires: 3600,
	})

}

func (s *qiniuStore) getReplaceToken(key string) string {
	return s.cli.MakeUptoken(&kodo.PutPolicy{
		Scope: s.bucket + ":" + key,
		//设置Token过期时间
		Expires: 3600,
	})

}
func errCodeMatch(err error, code int) bool {
	switch err.(type) {
	case *rpc.ErrorInfo:
		rei := err.(*rpc.ErrorInfo)
		return rei.Code == code
	default:
		return false
	}
}

func (s *qiniuStore) GetImageInfo(key string) (ii *store.ImageInfo, err error) {
	resp, err := http.Get(s.GetVisitURL(fmt.Sprintf("%s%s", key, imageInfoStyle)))
	if err != nil {
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		var res struct {
			Error string `json:"error"`
		}

		err = json.NewDecoder(resp.Body).Decode(&res)
		if err != nil {
			return
		}
		err = errors.New(res.Error)
		return
	}

	var iis store.ImageInfo
	err = json.NewDecoder(resp.Body).Decode(&iis)
	if err != nil {
		return
	}
	ii = &iis

	return
}

func (s *qiniuStore) SaveReaderAt(filename string, data io.ReaderAt, size int64) (err error) {
	token := s.getUploadToken()
	return s._saveReaderAt(filename, data, size, token)
}

func (s *qiniuStore) _saveReaderAt(filename string, data io.ReaderAt, size int64, token string) (err error) {
	zone := 0
	uploader := kodocli.NewUploader(zone, nil)
	var ret kodocli.PutRet

	return uploader.Rput(reqid.NewContext(context.Background(), filename), &ret, token, filename, data, size, &rputExtra)

}

// 外部文件一起打包
// packfile 返回打包文件路径
func (s *qiniuStore) ExternalMultifilePackaging(w io.Writer, externalFiles []store.ExternalFileAlias, keys ...store.FileAlias) error {
	//buffer := new(bytes.Buffer)
	writer := zip.NewWriter(w)
	defer writer.Close()
	var errInfo error
	for _, file := range keys {
		wz, err := writer.CreateHeader(&zip.FileHeader{
			Name:   file.Alias,
			Flags:  1 << 11,
			Method: zip.Deflate,
		})
		if err != nil {
			errInfo = err
			break
		}
		f, err := s.getReader(file.Key)
		if err != nil {
			errInfo = err
			break
		}

		io.Copy(wz, f)
		f.Close()
	}
	if errInfo != nil {
		return errInfo
	}
	// 处理外部文件
	for _, externalFile := range externalFiles {
		wz, err := writer.CreateHeader(&zip.FileHeader{
			Name:   externalFile.Alias,
			Flags:  1 << 11,
			Method: zip.Deflate,
		})
		if err != nil {
			errInfo = err
			break
		}
		io.Copy(wz, externalFile.FileRead)
	}
	return errInfo
}

// 外部文件一起打包，返回打包数据
// packfile 返回打包文件路径
func (s *qiniuStore) ExternalMultifileOutZipPackage(externalFiles []store.ExternalFileAlias, keys ...store.FileAlias) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	writer := zip.NewWriter(buffer)
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
		f, err := s.getReader(file.Key)
		if err != nil {
			errInfo = err
			break
		}

		io.Copy(w, f)
		f.Close()
	}
	if errInfo != nil {
		return nil, errInfo
	}
	// 处理外部文件
	for _, externalFile := range externalFiles {
		w, err := writer.CreateHeader(&zip.FileHeader{
			Name:   externalFile.Alias,
			Flags:  1 << 11,
			Method: zip.Deflate,
		})
		if err != nil {
			errInfo = err
			break
		}
		io.Copy(w, externalFile.FileRead)
	}
	return buffer, errInfo
}
