package filesystem

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

type Metadata struct {
	ID         ID
	FileID     ID
	Mode       os.FileMode
	Name       string
	Attributes map[string]interface{}
}

func (metadata *Metadata) MarshalJSON() ([]byte, error) {
	type Alias Metadata
	return json.Marshal(&struct {
		Mode string
		*Alias
	}{
		Mode:  fmt.Sprintf("%#o", metadata.Mode),
		Alias: (*Alias)(metadata),
	})
}

func (metadata *Metadata) UnmarshalJSON(data []byte) error {
	type Alias Metadata
	aux := &struct {
		Mode string
		*Alias
	}{
		Alias: (*Alias)(metadata),
	}

	err := json.Unmarshal(data, &aux)
	if err == nil {
		var i int64
		i, err = strconv.ParseInt(aux.Mode, 8, 32)
		metadata.Mode = os.FileMode(i)
	}
	return err
}

type FileInfo struct {
	os.FileInfo
	*Metadata
}

func (fi *FileInfo) ID() ID            { return fi.Metadata.ID }
func (fi *FileInfo) IsDir() bool       { return fi.Metadata.Mode.IsDir() }
func (fi *FileInfo) Mode() os.FileMode { return fi.Metadata.Mode }
func (fi *FileInfo) Name() string      { return fi.Metadata.Name }
func (fi *FileInfo) Sys() interface{}  { return nil }
