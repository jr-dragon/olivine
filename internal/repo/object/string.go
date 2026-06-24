package object

import "time"

type String struct {
	base

	val string
}

func NewString(k, v string, expiresAt *time.Time) *String {
	return &String{
		val:  v,
		base: base{key: k, expiresAt: expiresAt},
	}
}

func (str *String) String() string {
	return str.val
}
