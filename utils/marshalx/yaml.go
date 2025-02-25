package marshalx

import (
	"gopkg.in/yaml.v3"

	"github.com/go-xuan/quanx/os/errorx"
	"github.com/go-xuan/quanx/os/filex"
)

type Yaml struct{}

func (s Yaml) Name() string {
	return yamlStrategy
}

func (s Yaml) Marshal(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

func (s Yaml) Unmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}

// WriteYaml 写入yaml文件
func WriteYaml(path string, v any) error {
	bytes, err := yaml.Marshal(v)
	if err != nil {
		return errorx.Wrap(err, "yamlStrategy marshal error")
	}
	if err = filex.WriteFile(path, bytes); err != nil {
		return errorx.Wrap(err, "write file error")
	}
	return nil
}

func (s Yaml) Read(path string, v interface{}) error {
	if !filex.Exists(path) {
		return errorx.Errorf("the file not exist: %s", path)
	} else if data, err := filex.ReadFile(path); err != nil {
		return errorx.Wrap(err, "read file error")
	} else {
		return s.Unmarshal(data, v)
	}
}

func (s Yaml) Write(path string, v interface{}) error {
	if data, err := s.Marshal(v); err != nil {
		return errorx.Wrap(err, "yaml marshal error")
	} else if err = filex.WriteFile(path, data); err != nil {
		return errorx.Wrap(err, "write file error")
	}
	return nil
}
