package codex

import (
	"github.com/quanxiaoxuan/quanx/common/constx"
	"strings"

	"github.com/quanxiaoxuan/quanx/utils/sqlx"
	"github.com/quanxiaoxuan/quanx/utils/stringx"
)

// 构建go结构体代码
func BuildGoStruct(table string, fieldList FieldList) string {
	table = strings.TrimPrefix(table, `t_`)
	table = strings.TrimSuffix(table, `_t`)
	table = stringx.UpperCamelCase(table)
	sb := strings.Builder{}
	sb.WriteString(constx.NextLine)
	sb.WriteString("type ")
	sb.WriteString(table)
	sb.WriteString("List []*")
	sb.WriteString(table)
	sb.WriteString(constx.NextLine)
	sb.WriteString("type ")
	sb.WriteString(table)
	sb.WriteString(" struct {")
	for _, field := range fieldList {
		gt := sqlx.Pg2GoTypeMap[field.Type]
		up := stringx.UpperCamelCase(field.Name)
		low := stringx.LowerCamelCase(field.Name)
		sb.WriteString(constx.NextLine)
		sb.WriteString(constx.Tab)
		sb.WriteString(up)
		sb.WriteString(" ")
		sb.WriteString(gt)
		sb.WriteString(" `json:\"")
		sb.WriteString(low)
		sb.WriteString("\"` // ")
		sb.WriteString(field.Comment)
	}
	sb.WriteString(constx.NextLine)
	sb.WriteString("}")
	sb.WriteString(constx.NextLine)
	return sb.String()
}

// 构建go结构体代码
func BuildGormStruct(table string, fieldList FieldList) string {
	table = strings.TrimPrefix(table, `t_`)
	table = strings.TrimSuffix(table, `_t`)
	table = stringx.UpperCamelCase(table)
	sb := strings.Builder{}
	sb.WriteString(constx.NextLine)
	sb.WriteString("type ")
	sb.WriteString(table)
	sb.WriteString("List []*")
	sb.WriteString(table)
	sb.WriteString(constx.NextLine)
	sb.WriteString("type ")
	sb.WriteString(table)
	sb.WriteString(" struct {")
	for _, field := range fieldList {
		goType := sqlx.Pg2GoTypeMap[field.Type]
		gormType := sqlx.Pg2GormTypeMap[field.Type]
		up := stringx.UpperCamelCase(field.Name)
		low := stringx.LowerCamelCase(field.Name)
		sb.WriteString(constx.NextLine)
		sb.WriteString(constx.Tab)
		sb.WriteString(up)
		sb.WriteString(" ")
		sb.WriteString(goType)
		sb.WriteString(" `json:\"")
		sb.WriteString(low)
		sb.WriteString("\" gorm:\"type:")
		sb.WriteString(gormType)
		sb.WriteString("; comment:")
		sb.WriteString(field.Comment)
		sb.WriteString(";")
		sb.WriteString("\"`")
	}
	sb.WriteString(constx.NextLine)
	sb.WriteString("}")
	sb.WriteString(constx.NextLine)
	return sb.String()
}
