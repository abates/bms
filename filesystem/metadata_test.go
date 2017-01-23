package filesystem

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func TestMarshal(t *testing.T) {
	tests := []struct {
		id         ID
		mode       os.FileMode
		modeString string
	}{
		{id("6ba7b810-9dad-11d1-80b4-00c04fd430c8"), 0700, "0700"},
		{id("6ba7b810-9dad-11d1-80b4-00c04fd430c8"), 0600, "0600"},
	}

	for i, test := range tests {
		metadata := &Metadata{
			ID:   test.id,
			Mode: test.mode,
		}

		data, _ := metadata.MarshalJSON()
		dec := json.NewDecoder(bytes.NewBuffer(data))
		var v map[string]interface{}
		dec.Decode(&v)
		if modeString := v["Mode"].(string); modeString != test.modeString {
			t.Errorf("Test %d failed.  Expected '%s' got '%s'", i, test.modeString, modeString)
		}

		metadata = &Metadata{}
		metadata.UnmarshalJSON(data)

		if test.mode != metadata.Mode {
			t.Errorf("Test %d failed. Expected %v got %v", i, test.mode, metadata.Mode)
		}
	}
}

func TestFileInfo(t *testing.T) {
	tests := []struct {
		id   ID
		dir  bool
		mode os.FileMode
		name string
	}{
		{id("6ba7b810-9dad-11d1-80b4-00c04fd430c8"), false, os.ModePerm, "foo"},
		{id("6ba7b810-9dad-11d1-80b4-00c04fd430c8"), true, os.ModePerm, "bar"},
	}

	for i, test := range tests {
		if test.dir {
			test.mode = test.mode | os.ModeDir
		}

		metadata := &Metadata{
			ID:   test.id,
			Name: test.name,
			Mode: test.mode,
		}

		fi := &FileInfo{nil, metadata}
		if !bytes.Equal(test.id[:], fi.ID()[:]) {
			t.Errorf("Test %d failed.  Expected ID %s got %s", i, test.id.String(), fi.ID().String())
		}

		if test.dir != fi.IsDir() {
			t.Errorf("Test %d failed.  Expected IsDir() to be %v got %v", i, test.dir, fi.IsDir())
		}

		if test.mode != fi.Mode() {
			t.Errorf("Test %d failed.  Expected mode to be %v got %v", i, test.mode, fi.Mode())
		}

		if test.name != fi.Name() {
			t.Errorf("Test %d failed.  Expected name to be %s got %s", i, test.name, fi.Name())
		}

		if fi.Sys() != nil {
			t.Errorf("Test %d failed.  Expected Sys() to be nil", i)
		}
	}
}
