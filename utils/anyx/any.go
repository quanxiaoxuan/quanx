package anyx

import (
	"reflect"
)

// if取值
func If[T any](x bool, t, f T) T {
	if x {
		return t
	} else {
		return f
	}
}

// 判空时取默认值
func IfZero[T any](x, def T) T {
	if reflect.ValueOf(&x).Elem().IsZero() {
		return def
	}
	return x
}

func Default[T any](x []T, def T) T {
	if len(x) == 0 {
		return def
	} else {
		return x[0]
	}
}
