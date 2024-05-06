package gatewayx

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-xuan/quanx/core/ginx"
	"github.com/go-xuan/quanx/core/nacosx"
	"github.com/go-xuan/quanx/types/stringx"
	"github.com/go-xuan/quanx/utils/marshalx"
)

var Servers []*Server

// 微服务网关配置
type Server struct {
	Name     string   `yaml:"name" json:"name"`         // 微服务名称
	Group    string   `yaml:"group" json:"group"`       // 微服务分组
	Prefix   string   `yaml:"prefix" json:"prefix"`     // 微服务API前缀
	Router   string   `yaml:"router" json:"router"`     // API路由
	AuthType string   `yaml:"authType" json:"authType"` // 鉴权方式(cookie/token/no)
	Ignore   []string `yaml:"ignore" json:"ignore"`     // 忽略鉴权(白名单)
}

// 获取微服务addr
func GetServerProxyAddr(group, dataId, url string) (addr string, authType string, err error) {
	if err = ListenConfigChanged(group, dataId); err != nil {
		err = errors.New("监听微服务网关配置失败 ：" + err.Error())
		return
	}
	for _, server := range Servers {
		if MatchUrl(url, server.Router) {
			authType = server.AuthType
			if authType != ginx.NoAuth && len(server.Ignore) > 0 {
				for _, item := range server.Ignore {
					if stringx.Index(url, strings.TrimSpace(item)) >= 0 {
						authType = ginx.NoAuth
					}
				}
			}
			if addr, err = nacosx.SelectOneHealthyInstance(server.Name, server.Group); err != nil {
				err = errors.New("微服务实例未注册 ：" + err.Error())
				return
			}
			addr = "http://" + addr + server.Prefix
			return
		}
	}
	err = fmt.Errorf("未找到对应的网关路由配置，请检查微服务配置文件，或者确认请求接口[%s]是否正确", url)
	return

}

// 校验nacos配置
func ListenConfigChanged(group, dataId string) (err error) {
	if data, ok := nacosx.GetNacosConfigMonitor().GetConfigData(group, dataId); ok && data.Changed {
		// 将当前最新的content数据同步到servers
		if err = marshalx.NewCase(dataId).Unmarshal([]byte(data.Content), &Servers); err != nil {
			return
		}
		// 更新nacos监控中配置值
		data.Changed = false
		data.UpdateTime = time.Now().UnixMilli()
	}
	return
}

// URL匹配
func MatchUrl(uri, rule string) bool {
	if rule == "*" || rule == "/*" {
		return true
	} else if stringx.Index(rule, ",") >= 0 {
		var list = strings.Split(rule, ",")
		for _, item := range list {
			if stringx.Index(uri, strings.TrimSpace(item)) >= 0 {
				return true
			}
		}
	} else if strings.HasSuffix(rule, `/**`) {
		return strings.HasPrefix(uri, strings.TrimSuffix(rule, `/**`))
	} else if strings.HasSuffix(rule, `/*`) {
		var prefix = strings.TrimSuffix(rule, `/*`)
		if strings.HasPrefix(uri, prefix) {
			uri = uri[len(prefix):]
			return stringx.Index(uri, `/`) < 0
		}
	} else {
		return stringx.Index(uri, rule) >= 0
	}
	return false
}