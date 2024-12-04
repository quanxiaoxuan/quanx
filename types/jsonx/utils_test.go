package jsonx

import (
	"fmt"
	"github.com/tidwall/gjson"
	"testing"
)

func TestJsonGet(t *testing.T) {
	j := `
		{
		  "name": {"first": "Tom", "last": "Anderson"},
		  "age":37,
		  "children": ["Sara","Alex","Jack"],
		  "friends": [
		    {"first": "James", "last": "Murphy"},
		    {"first": "Roger", "last": "Craig"}
		  ]
		}`
	fmt.Println(gjson.Get(j, "friends.1").Value())
	fmt.Println(Get(j, "friends.[1].first").Value())
}