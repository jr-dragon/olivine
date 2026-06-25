package object

import "time"

type Object interface {
	Key() string
	ExpiresAt() *time.Time
	Expired() bool
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

func (obj *base) Expired() bool {
	if obj.expiresAt == nil {
		return false
	}

	return time.Now().After(*obj.expiresAt)
}

func (obj *base) SetExpiresAt(exp *time.Time) {
	obj.expiresAt = exp
}
