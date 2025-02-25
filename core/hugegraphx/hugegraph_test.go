package hugegraphx

import (
	"testing"
	
	"github.com/go-xuan/quanx/core/configx"
)

func TestHugegraph(t *testing.T) {
	if err := configx.Execute(&Config{
		Host:  "localhost",
		Port:  8882,
		Graph: "hugegraph",
	}); err != nil {
		t.Error(err)
	}
}
