package object

import "time"

type Object interface {
	Key() string
	ExpiresAt() *time.Time
	SetExpiresAt(*time.Time)
}

type base struct {
	key       string
	expiresAt *time.Time
}

func (obj *base) Key() string {
	return obj.key
}

func (obj *base) ExpiresAt() *time.Time {
	return obj.expiresAt
}

func (obj *base) SetExpiresAt(exp *time.Time) {
	obj.expiresAt = exp
}
