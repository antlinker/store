package store_test

import "testing"
import . "github.com/antlinker/store"

func TestFileUpdate(t *testing.T) {
	InitFileStore()
	testUpdate(t)
	testRead(t)
	testMultifilePackaging(t)
}
