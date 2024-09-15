package internal

import "unsafe"

func StrPtr(s string) *uint8 {
	return unsafe.StringData(s)
}
