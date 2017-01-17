package file_test

import "testing"

import "github.com/antlinker/store/file"

func TestFileUpdate(t *testing.T) {
	file.InitStore("")
	testUpdate(t)
	testRead(t)
	testMultifilePackaging(t)
}
