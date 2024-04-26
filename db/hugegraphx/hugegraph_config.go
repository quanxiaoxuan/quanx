package hugegraphx

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/go-xuan/quanx/net/httpx"
	"github.com/go-xuan/quanx/server/confx"
)

// hugegraph配置
type Hugegraph struct {
	Host  string `json:"host" yaml:"host" nacos:"hugegraph.host"`    // 主机
	Port  int    `json:"port" yaml:"port" nacos:"hugegraph.port"`    // 端口
	Graph string `json:"graph" yaml:"graph" nacos:"hugegraph.graph"` // 图名称
}

// 配置信息格式化
func (h *Hugegraph) ToString(title string) string {
	return fmt.Sprintf("%s => host=%s port=%d graph=%s", title, h.Host, h.Port, h.Graph)
}

// 配置器名称
func (*Hugegraph) Theme() string {
	return "Hugegraph"
}

// 配置文件读取
func (*Hugegraph) Reader() *confx.Reader {
	return &confx.Reader{
		FilePath:    "hugegraph.yaml",
		NacosDataId: "hugegraph.yaml",
		Listen:      false,
	}
}

// 配置器运行
func (h *Hugegraph) Run() error {
	if h.Host == "" {
		return nil
	}
	if handler == nil {
		if h.Ping() {
			handler = &Handler{Config: h, GremlinUrl: h.GremlinUrl(), SchemaUrl: h.SchemaUrl()}
			log.Info(h.ToString("Hugegraph connect successful!"))
		} else {
			log.Error(h.ToString("Hugegraph connect failed!"))
		}
	}
	return nil
}

func (h *Hugegraph) GremlinUrl() string {
	return fmt.Sprintf("http://%s:%d/gremlin", h.Host, h.Port)
}

func (h *Hugegraph) SchemaUrl() string {
	return fmt.Sprintf("http://%s:%d/graphs/%s/schema/", h.Host, h.Port, h.Graph)
}

// gremlin查询API-get请求
func (h *Hugegraph) Ping() bool {
	if bytes, err := httpx.Get().Url(fmt.Sprintf("http://%s:%d/versions", h.Host, h.Port)).Do(); err != nil {
		return false
	} else {
		var resp = &PingResp{}
		if err = json.Unmarshal(bytes, &resp); err != nil {
			return false
		}
		if resp.Versions != nil && resp.Versions.Version != "" {
			return true
		}
		return false
	}
}
