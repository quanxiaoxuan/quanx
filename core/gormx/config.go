package gormx

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/go-xuan/quanx/common/constx"
	"github.com/go-xuan/quanx/core/configx"
	"github.com/go-xuan/quanx/os/errorx"
	"github.com/go-xuan/quanx/types/anyx"
)

func NewConfigurator(conf *Config) configx.Configurator {
	return conf
}

type Config struct {
	Source          string `json:"source" yaml:"source" default:"default"`              // 数据源名称
	Enable          bool   `json:"enable" yaml:"enable"`                                // 数据源启用
	Type            string `json:"type" yaml:"type"`                                    // 数据库类型
	Host            string `json:"host" yaml:"host" default:"localhost"`                // 数据库Host
	Port            int    `json:"port" yaml:"port"`                                    // 数据库端口
	Username        string `json:"username" yaml:"username"`                            // 用户名
	Password        string `json:"password" yaml:"password"`                            // 密码
	Database        string `json:"database" yaml:"database"`                            // 数据库名
	Schema          string `json:"schema" yaml:"schema"`                                // schema模式名
	Debug           bool   `json:"debug" yaml:"debug" default:"false"`                  // 开启debug（打印SQL以及初始化模型建表）
	MaxIdleConns    int    `json:"maxIdleConns" yaml:"maxIdleConns" default:"10"`       // 最大空闲连接
	MaxOpenConns    int    `json:"maxOpenConns" yaml:"maxOpenConns" default:"10"`       // 最大打开连接
	ConnMaxLifetime int    `json:"connMaxLifetime" yaml:"connMaxLifetime" default:"10"` // 连接存活时间(分钟)
}

func (c *Config) Format() string {
	return fmt.Sprintf("source=%s type=%s host=%s port=%v database=%s debug=%v",
		c.Source, c.Type, c.Host, c.Port, c.Database, c.Debug)
}

func (c *Config) Reader() *configx.Reader {
	return &configx.Reader{
		FilePath:    "database.yaml",
		NacosDataId: "database.yaml",
		Listen:      false,
	}
}

func (c *Config) Execute() error {
	if c.Enable {
		if err := anyx.SetDefaultValue(c); err != nil {
			return errorx.Wrap(err, "set default value error")
		}
		if db, err := c.NewGormDB(); err != nil {
			return errorx.Wrap(err, "new gorm.DB failed")
		} else {
			if _handler == nil {
				_handler = &Handler{
					multi:     false,
					config:    c,
					configMap: make(map[string]*Config),
					gormDB:    db,
					gormMap:   map[string]*gorm.DB{},
				}
			} else {
				_handler.multi = true
			}
			_handler.gormMap[c.Source] = db
			_handler.configMap[c.Source] = c
			return nil
		}
	}
	log.Info("database not connected! reason: database.yaml is empty or the value of enable is false")
	return nil
}

// NewGormDB 创建数据库连接
func (c *Config) NewGormDB() (*gorm.DB, error) {
	if db, err := c.GetGormDB(); err != nil {
		return nil, errorx.Wrap(err, "new gorm db failed")
	} else {
		var sqlDB *sql.DB
		if sqlDB, err = db.DB(); err != nil {
			return nil, errorx.Wrap(err, "get sql db failed")
		}
		sqlDB.SetMaxIdleConns(c.MaxIdleConns)
		sqlDB.SetMaxOpenConns(c.MaxOpenConns)
		sqlDB.SetConnMaxLifetime(time.Duration(c.ConnMaxLifetime) * time.Second)
		// 是否打印SQL
		if c.Debug {
			db = db.Debug()
		}
		return db, nil
	}
}

// 数据库类型
const (
	MYSQL    = "mysql"
	POSTGRES = "postgres"
	PGSQL    = "pgsql"
)

// CommentTableSql 生成表备注
func (c *Config) CommentTableSql(table, comment string) string {
	switch strings.ToLower(c.Type) {
	case MYSQL:
		return "alter table " + table + " comment = '" + comment + "'"
	case POSTGRES, PGSQL:
		return "comment on table " + table + " is '" + comment + "'"
	}
	return ""
}

// GetGormDB 根据dsn生成gormDB
func (c *Config) GetGormDB() (*gorm.DB, error) {
	var dial gorm.Dialector
	switch strings.ToLower(c.Type) {
	case MYSQL:
		dial = mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?clientFoundRows=false&parseTime=true&timeout=1800s&charset=utf8&collation=utf8_general_ci&loc=Local",
			c.Username, c.Password, c.Host, c.Port, c.Database))
	case POSTGRES, PGSQL:
		dial = postgres.Open(fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
			c.Host, c.Port, c.Username, c.Password, c.Database))
	default:
		return nil, errorx.Errorf("database type only support : %s or %s", MYSQL, POSTGRES)
	}
	if db, err := gorm.Open(dial, &gorm.Config{
		// 模型命名策略
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 表名单数命名
		},
	}); err != nil {
		return nil, errorx.Wrap(err, "gorm open dialector failed")
	} else {
		return db, nil
	}
}

// MultiConfig 数据库多数据源配置
type MultiConfig []*Config

func (m MultiConfig) Format() string {
	sb := &strings.Builder{}
	sb.WriteString("[")
	for i, config := range m {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("{")
		sb.WriteString(config.Format())
		sb.WriteString("}")
	}
	sb.WriteString("]")
	return sb.String()
}

func (MultiConfig) Reader() *configx.Reader {
	return &configx.Reader{
		FilePath:    "database.yaml",
		NacosDataId: "database.yaml",
		Listen:      false,
	}
}

func (m MultiConfig) Execute() error {
	if len(m) == 0 {
		return errorx.New("database not connected! cause: database.yaml is invalid")
	}
	if _handler == nil {
		_handler = &Handler{
			multi:     true,
			gormMap:   make(map[string]*gorm.DB),
			configMap: make(map[string]*Config),
		}
	} else {
		_handler.multi = true
	}
	for i, c := range m {
		if c.Enable {
			if err := anyx.SetDefaultValue(c); err != nil {
				return errorx.Wrap(err, "set default value error")
			}
			var db, err = c.NewGormDB()
			if err != nil {
				return errorx.Wrap(err, "new gorm.Config failed")
			}
			_handler.gormMap[c.Source] = db
			_handler.configMap[c.Source] = c
			if i == 0 || c.Source == constx.DefaultSource {
				_handler.gormDB = db
				_handler.config = c
			}
		}
	}
	if len(_handler.configMap) == 0 {
		log.Error("database not connected! cause: database.yaml is empty or no enabled database configured")
	}
	return nil
}
