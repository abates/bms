package filesystem

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestPathForID(t *testing.T) {
	tests := []struct {
		id   ID
		path string
	}{
		{id("6ba7b810-9dad-11d1-80b4-00c04fd430c8"), "/NOT3/QEE5/VUI5/DAFU/ADAE/7VBQ/ZA"},
	}

	for i, test := range tests {
		path := pathForID(test.id)
		if test.path != path {
			t.Errorf("Test %d: Expected %s got %s", i, test.path, path)
		}
	}
}

func TestOpen(t *testing.T) {
	tests := []struct {
		id   ID
		flag int
		err  error
	}{
		{nil, os.O_RDWR | os.O_CREATE, nil},
		{nil, os.O_RDWR, &os.PathError{"OpenFile", "", fmt.Errorf("invalid id")}},
		{id("6ba7b810-9dad-11d1-80b4-00c04fd430c8"), os.O_RDWR | os.O_CREATE, nil},
		{id("6ba7b810-9dad-11d1-80b4-00c04fd430c8"), os.O_RDWR, nil},
	}

	testfs := newTestfs()
	for i, test := range tests {
		basefs := NewBaseFileSystem(testfs)

		var asset Asset
		var err error
		if test.id == nil {
			asset, err = basefs.OpenAssetByID(test.id, test.flag, 0006)
		} else {
			asset, err = basefs.OpenAsset(test.id.String(), test.flag, 0006)
		}

		if test.err == nil {
			if err != nil {
				t.Errorf("Test %d failed: %v", i, err)
			} else {
				// check to make sure initial metadata asset was written
				_, err = testfs.Stat(pathForID(asset.Metadata().ID))
				if err != nil {
					t.Errorf("Test %d failed: %v", i, err)
				}

				// check to make sure actual asset was created
				_, err = testfs.Stat(pathForID(asset.Metadata().FileID))
				if err != nil {
					t.Errorf("Test %d failed: %v", i, err)
				}
			}
		} else {
			if !reflect.DeepEqual(test.err, err) {
				t.Errorf("Test %d expected error '%v' but got '%v'", i, test.err, err)
			}
		}
	}
}

func TestUnsupported(t *testing.T) {
	basefs := NewBaseFileSystem(newTestfs())
	err := basefs.Mkdir("", 0)
	if err != ErrNotSupported {
		t.Errorf("Expected Mkdir to return '%v' but got '%v'", ErrNotSupported, err)
	}

	err = basefs.MkdirAll("", 0)
	if err != ErrNotSupported {
		t.Errorf("Expected MkdirAll to return '%v' but got '%v'", ErrNotSupported, err)
	}

	err = basefs.Rename("", "")
	if err != ErrNotSupported {
		t.Errorf("Expected Rename to return '%v' but got '%v'", ErrNotSupported, err)
	}
}

func TestStat(t *testing.T) {
	tests := []struct {
		id   string
		mode os.FileMode
	}{
		{"6ba7b810-9dad-11d1-80b4-00c04fd430c8", 0700},
		{"cd7cb54d-ae8f-4f6f-8f11-3ba26b493538", 0770},
		{"87c13e63-1750-4194-b39d-1f9b8def5db5", 0777},
	}

	basefs := NewBaseFileSystem(newTestfs())
	for i, test := range tests {
		asset, _ := basefs.OpenAsset(test.id, os.O_CREATE, test.mode)
		asset.Close()

		fi, err := basefs.Stat(test.id)
		if err != nil {
			t.Errorf("Test %d failed: %v", i, err)
		}

		if test.mode != fi.Mode() {
			t.Errorf("Test %d failed.  Expected mode %s got %s", i, test.mode, fi.Mode())
		}
	}
}
