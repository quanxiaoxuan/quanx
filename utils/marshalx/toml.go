package marshalx

import (
	"bytes"
	"github.com/BurntSushi/toml"

	"github.com/go-xuan/quanx/os/errorx"
	"github.com/go-xuan/quanx/os/filex"
)

type Toml struct{}

func (t Toml) Name() string {
	return tomlMethod
}

func (t Toml) Marshal(v interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	if err := toml.NewEncoder(&buffer).Encode(v); err != nil {
		return nil, errorx.Wrap(err, "encode toml failed")
	}
	return buffer.Bytes(), nil
}

func (t Toml) Unmarshal(data []byte, v interface{}) error {
	return toml.Unmarshal(data, v)
}

func (t Toml) Read(path string, v interface{}) error {
	if !filex.Exists(path) {
		return errorx.Errorf("the file not exist: %t", path)
	} else if data, err := filex.ReadFile(path); err != nil {
		return errorx.Wrap(err, "read file error")
	} else {
		return t.Unmarshal(data, v)
	}
}

func (t Toml) Write(path string, v interface{}) error {
	if data, err := t.Marshal(v); err != nil {
		return errorx.Wrap(err, "toml marshal error")
	} else if err = filex.WriteFile(path, data); err != nil {
		return errorx.Wrap(err, "write file error")
	}
	return nil
}
