package nacosx

import (
	"path/filepath"
	"strings"

	"github.com/go-xuan/quanx/app/constx"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
	log "github.com/sirupsen/logrus"

	"github.com/go-xuan/quanx/app/configx"
	"github.com/go-xuan/quanx/types/stringx"
	"github.com/go-xuan/quanx/utils/fmtx"
)

const (
	OnlyConfig      = iota // 仅用配置中心
	OnlyNaming             // 仅用服务发现
	ConfigAndNaming        // 配置中心和服务发现都使用
)

// Nacos nacos访问配置
type Nacos struct {
	Address   string `yaml:"address" json:"address" default:"127.0.0.1"`  // nacos服务地址,多个以英文逗号分割
	Username  string `yaml:"username" json:"username" default:"nacos"`    // 用户名
	Password  string `yaml:"password" json:"password" default:"nacos"`    // 密码
	NameSpace string `yaml:"nameSpace" json:"nameSpace" default:"public"` // 命名空间
	Mode      int    `yaml:"mode" json:"mode" default:"2"`                // 模式（0-仅配置中心；1-仅服务发现；2-配置中心和服务发现）
}

func (n *Nacos) ID() string {
	return "nacos"
}

func (n *Nacos) Format() string {
	return fmtx.Yellow.XSPrintf("address=%s username=%s password=%s nameSpace=%s mode=%v",
		n.AddressUrl(), n.Username, n.Password, n.NameSpace, n.Mode)
}

func (*Nacos) Reader() *configx.Reader {
	return nil
}

func (n *Nacos) Execute() (err error) {
	if handler == nil {
		handler = &Handler{Config: n}
		switch n.Mode {
		case OnlyConfig:
			if handler.ConfigClient, err = n.ConfigClient(n.ClientParam()); err != nil {
				return
			}
		case OnlyNaming:
			if handler.NamingClient, err = n.NamingClient(n.ClientParam()); err != nil {
				return
			}
		case ConfigAndNaming:
			var param = n.ClientParam()
			if handler.ConfigClient, err = n.ConfigClient(param); err != nil {
				return
			}
			if handler.NamingClient, err = n.NamingClient(param); err != nil {
				return
			}
		}
	}
	log.Info("nacos connect successfully: ", n.Format())
	return
}

// AddressUrl nacos访问地址
func (n *Nacos) AddressUrl() string {
	return n.Address + "/nacos"
}

// EnableNaming 开启服务注册
func (n *Nacos) EnableNaming() bool {
	return n.Mode == OnlyNaming || n.Mode == ConfigAndNaming
}

// ClientConfig nacos客户端配置
func (n *Nacos) ClientConfig() *constant.ClientConfig {
	return &constant.ClientConfig{
		Username:            n.Username,
		Password:            n.Password,
		TimeoutMs:           10 * 1000,
		BeatInterval:        3 * 1000,
		NotLoadCacheAtStart: true,
		NamespaceId:         n.NameSpace,
		LogDir:              filepath.Join(constx.DefaultResourceDir, ".nacos/log"),
		CacheDir:            filepath.Join(constx.DefaultResourceDir, ".nacos/cache"),
	}
}

// ServerConfigs nacos服务中间件配置
func (n *Nacos) ServerConfigs() []constant.ServerConfig {
	var adds = strings.Split(n.Address, ",")
	if len(adds) == 0 {
		log.Error("the address of nacos cannot be empty")
		return nil
	}
	var configs []constant.ServerConfig
	for _, addStr := range adds {
		host, port, _ := strings.Cut(addStr, ":")
		configs = append(configs, constant.ServerConfig{
			ContextPath: "/nacos",
			IpAddr:      host,
			Port:        uint64(stringx.ToInt64(port)),
		})
	}
	return configs
}

func (n *Nacos) ClientParam() vo.NacosClientParam {
	return vo.NacosClientParam{
		ClientConfig:  n.ClientConfig(),
		ServerConfigs: n.ServerConfigs(),
	}
}

// ConfigClient 初始化Nacos配置中心客户端
func (n *Nacos) ConfigClient(param vo.NacosClientParam) (client config_client.IConfigClient, err error) {
	if client, err = clients.NewConfigClient(param); err != nil {
		log.Error("Nacos Config Client Init Failed: ", n.Format())
		log.Error(err)
		return
	}
	return
}

// NamingClient 初始化Nacos服务发现客户端
func (n *Nacos) NamingClient(param vo.NacosClientParam) (client naming_client.INamingClient, err error) {
	if client, err = clients.NewNamingClient(param); err != nil {
		log.Error("Nacos Naming Client Init Failed: ", n.Format(), err)
		return
	}
	return
}

// Register 初始化Nacos服务发现客户端
func (n *Nacos) Register(server ServerInstance) (err error) {
	var client naming_client.INamingClient
	if client = This().NamingClient; client == nil {
		if client, err = n.NamingClient(n.ClientParam()); err != nil {
			return
		}
	}
	if _, err = client.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          server.Host,
		Port:        uint64(server.Port),
		GroupName:   n.NameSpace,
		ServiceName: server.Name,
		Weight:      1,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    nil,
	}); err != nil {
		log.Error("nacos server register Failed: ", server.Info(), err)
	} else {
		log.Info("Nacos server register successfully: ", server.Info())
	}
	return
}

func (n *Nacos) Deregister(server ServerInstance) (err error) {
	var client naming_client.INamingClient
	if client = This().NamingClient; client == nil {
		if client, err = n.NamingClient(n.ClientParam()); err != nil {
			return
		}
	}
	if _, err = client.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          server.Host,
		Port:        uint64(server.Port),
		GroupName:   n.NameSpace,
		ServiceName: server.Name,
		Ephemeral:   true,
	}); err != nil {
		log.Error("nacos server deregister failed: ", server.Info(), err)
	} else {
		log.Info("nacos server deregister successfully: ", server.Info())
	}
	return
}

func (n *Nacos) AllInstances(name string) (instances []*ServerInstance, err error) {
	var servers []model.Instance
	if servers, err = This().NamingClient.SelectInstances(vo.SelectInstancesParam{
		ServiceName: name,
		GroupName:   n.NameSpace,
		HealthyOnly: true,
	}); err != nil {
		return
	}
	for _, server := range servers {
		instances = append(instances, &ServerInstance{
			Name: server.ServiceName,
			Host: server.Ip,
			Port: int(server.Port),
		})
	}
	return
}

func (n *Nacos) GetInstance(name string) (instance *ServerInstance, err error) {
	var servers *model.Instance
	if servers, err = This().NamingClient.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: name,
		GroupName:   n.NameSpace,
	}); err != nil {
		return
	}
	instance = &ServerInstance{
		Name: servers.ServiceName,
		Host: servers.Ip,
		Port: int(servers.Port),
	}
	return
}
