package codex

import (
	"github.com/go-xuan/quanx/public/constx"
	"strings"

	"github.com/go-xuan/quanx/utils/sqlx"
	"github.com/go-xuan/quanx/utils/stringx"
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
	sb.WriteString(" struct {")
	for _, field := range fieldList {
		sb.WriteString(constx.NextLine)
		sb.WriteString(constx.Tab)
		sb.WriteString(stringx.UpperCamelCase(field.Name))
		sb.WriteString(" ")
		sb.WriteString(sqlx.Pg2GoTypeMap[field.Type])
		sb.WriteString(" `json:\"")
		sb.WriteString(stringx.LowerCamelCase(field.Name))
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
	sb.WriteString(" struct {")
	for _, field := range fieldList {
		sb.WriteString(constx.NextLine)
		sb.WriteString(constx.Tab)
		sb.WriteString(stringx.UpperCamelCase(field.Name))
		sb.WriteString(" ")
		sb.WriteString(sqlx.Pg2GoTypeMap[field.Type])
		sb.WriteString(" `json:\"")
		sb.WriteString(stringx.LowerCamelCase(field.Name))
		sb.WriteString("\" gorm:\"type:")
		sb.WriteString(sqlx.Pg2GormTypeMap[field.Type])
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
