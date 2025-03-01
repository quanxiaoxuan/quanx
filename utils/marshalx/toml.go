package marshalx

import (
	"bytes"
	"github.com/BurntSushi/toml"

	"github.com/go-xuan/quanx/os/errorx"
	"github.com/go-xuan/quanx/os/filex"
)

type Toml struct{}

func (s Toml) Name() string {
	return tomlStrategy
}

func (s Toml) Marshal(v interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	if err := toml.NewEncoder(&buffer).Encode(v); err != nil {
		return nil, errorx.Wrap(err, "encode toml failed")
	}
	return buffer.Bytes(), nil
}

func (s Toml) Unmarshal(data []byte, v interface{}) error {
	return toml.Unmarshal(data, v)
}

func (s Toml) Read(path string, v interface{}) error {
	if !filex.Exists(path) {
		return errorx.Errorf("the file not exist: %s", path)
	} else if data, err := filex.ReadFile(path); err != nil {
		return errorx.Wrap(err, "read file error")
	} else {
		return s.Unmarshal(data, v)
	}
}

func (s Toml) Write(path string, v interface{}) error {
	if data, err := s.Marshal(v); err != nil {
		return errorx.Wrap(err, "toml marshal error")
	} else if err = filex.WriteFile(path, data); err != nil {
		return errorx.Wrap(err, "write file error")
	}
	return nil
}
