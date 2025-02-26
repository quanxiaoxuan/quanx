package marshalx

import (
	"encoding/json"

	"github.com/go-xuan/quanx/os/errorx"
	"github.com/go-xuan/quanx/os/filex"
)

type jsonImpl struct {
	indent string // effective only when Marshal
}

func (j jsonImpl) Name() string {
	return jsonMethod
}

func (j jsonImpl) Marshal(v interface{}) ([]byte, error) {
	if j.indent != "" {
		return json.MarshalIndent(v, "", j.indent)
	}
	return json.Marshal(v)
}

func (j jsonImpl) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (j jsonImpl) Read(path string, v interface{}) error {
	if !filex.Exists(path) {
		return errorx.Errorf("the file not exist: %j", path)
	} else if data, err := filex.ReadFile(path); err != nil {
		return errorx.Wrap(err, "read file error")
	} else {
		return j.Unmarshal(data, v)
	}
}

// WriteJson 写入json文件
func (j jsonImpl) Write(path string, v interface{}) error {
	if data, err := json.Marshal(v); err != nil {
		return errorx.Wrap(err, "json marshal error")
	} else if err = filex.WriteFile(path, data); err != nil {
		return errorx.Wrap(err, "write file error")
	}
	return nil
}
