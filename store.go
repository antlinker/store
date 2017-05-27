package store

import (
	"bytes"
	"io"
	"time"
)

// Storer 存储文件接口
type Storer interface {
	// SaveFile 将数据data存储到filenamet
	SaveData(filename string, data []byte) (err error)
	SaveReader(filename string, data io.Reader, size int64) (err error)
	SaveFile(filename string, srcfile string) (err error)
	// DeleteFile 删除指定文件
	DeleteFile(filename string) (err error)

	UpdateData(filename string, data []byte) error
	UpdateFile(filename string, srcfile string) (err error)
	UpdateReader(filename string, data io.Reader, size int64) (err error)
	IsExists(filename string) (bool, error)
	Fsize(filename string) (int64, error)
	PutTime(key string) (time.Time, error)
	Copy(src, dst string) error
	Move(src, dst string) error

	SetVisitHTTPBase(path string)
	GetVisitURL(key string) string
	// 获取文件流
	// 如果是文件系统存储时可能是自身
	// 如果是云存储时可能来自远程
	GetReader(key string) (io.ReadCloser, error)
	// 文件打包
	MultifilePackaging(w io.Writer, keys ...FileAlias) (err error)

	// 获取图片信息
	GetImageInfo(key string) (ii *ImageInfo, err error)

	SaveReaderAt(filename string, data io.ReaderAt, size int64) (err error)

	// // 外部文件一起打包
	// ExternalMultifilePackaging(w io.Writer, externalFiles []ExternalFileAlias, keys ...FileAlias) (err error)
	// // 输出压缩包
	// ExternalMultifileOutZipPackage(externalFiles []ExternalFileAlias, keys ...FileAlias) (buffer *bytes.Buffer, err error)
}

// ImageInfo 图片信息
type ImageInfo struct {
	Format      string `json:"format"`      // 图片类型，如pngjpeggifbmp等。
	Width       int    `json:"width"`       // 图片宽度，单位：像素(px)。
	Height      int    `json:"height"`      // 图片高度，单位：像素(px)。
	ColorModel  string `json:"colorModel"`  // 彩色空间，如palette16ycbcr等。
	FrameNumber int    `json:"frameNumber"` // 帧数，gif 图片会返回此项。
}

// FileAlias 文件别名映射
type FileAlias struct {
	Key   string
	Alias string
}

// // ExternalFileAlias 文件别名映射
// type ExternalFileAlias struct {
// 	FileRead io.Reader
// 	Alias    string
// }

// DefaultStore 默认存储
var DefaultStore Storer

// SaveFile 将数据data存储到filenamet
func SaveFile(filename string, srcfile string) (err error) {
	return DefaultStore.SaveFile(filename, srcfile)
}

// Copy 将源文件复制到目标文件
func Copy(src string, dst string) (err error) {
	return DefaultStore.Copy(src, dst)
}

// Move 将源文件移动到目标文件
func Move(src string, dst string) (err error) {
	return DefaultStore.Move(src, dst)
}

// UpdateFile 将数据data存储到filenamet
func UpdateFile(filename string, srcfile string) (err error) {
	return DefaultStore.UpdateFile(filename, srcfile)
}

// SaveData 将data存储到文件
func SaveData(filename string, data []byte) error {
	return DefaultStore.SaveData(filename, data)
}

// SaveReader 将data存储到文件
func SaveReader(filename string, data io.Reader, size int64) error {
	return DefaultStore.SaveReader(filename, data, size)
}

// UpdateData 将数据data存储到filenamet
func UpdateData(filename string, data []byte) error {
	return DefaultStore.UpdateData(filename, data)
}

// UpdateReader 将数据data存储到filenamet
func UpdateReader(filename string, data io.Reader, size int64) error {
	return DefaultStore.UpdateReader(filename, data, size)
}

// DeleteFile 删除指定文件
func DeleteFile(filename string) error {
	return DefaultStore.DeleteFile(filename)
}

// IsExists 判断文件是否存在
func IsExists(filename string) (bool, error) {
	return DefaultStore.IsExists(filename)
}

// PutTime 判断文件修改时间
func PutTime(filename string) (time.Time, error) {
	return DefaultStore.PutTime(filename)
}

// Fsize 获取文件大小
func Fsize(filename string) (int64, error) {
	return DefaultStore.Fsize(filename)
}

// SetVisitHTTPBase 设置基础路径
func SetVisitHTTPBase(path string) {
	DefaultStore.SetVisitHTTPBase(path)
}

// GetVisitURL 获取访问url
func GetVisitURL(key string) string {
	return DefaultStore.GetVisitURL(key)
}

// MultifilePackaging 多文件打包
func MultifilePackaging(w io.Writer, keys ...FileAlias) (err error) {
	return DefaultStore.MultifilePackaging(w, keys...)
}

// GetReader 获取文件流
func GetReader(key string) (io.ReadCloser, error) {
	return DefaultStore.GetReader(key)
}

// ExternalMultifilePackaging 外部文件与七牛云文件一起打包
func ExternalMultifilePackaging(w io.Writer, externalFiles []ExternalFileAlias, keys ...FileAlias) (err error) {
	return DefaultStore.ExternalMultifilePackaging(w, externalFiles, keys...)
}

// ExternalMultifileOutZipPackage 打包后返回数据
func ExternalMultifileOutZipPackage(externalFiles []ExternalFileAlias, keys ...FileAlias) (buffer *bytes.Buffer, err error) {
	return DefaultStore.ExternalMultifileOutZipPackage(externalFiles, keys...)
}
