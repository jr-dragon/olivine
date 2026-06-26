package object

import (
	"strconv"
	"time"

	"github.com/zeebo/xxh3"
)

type String struct {
	*base

	val    string
	digest uint64
}

func NewString(k, v string, expiresAt *time.Time) *String {
	return &String{
		val:    v,
		digest: xxh3.HashString(v),
		base:   &base{key: k, expiresAt: expiresAt},
	}
}

func (str *String) String() string {
	return str.val
}

func (str *String) Equals(v string) bool {
	return str.String() == v
}

func (str *String) EqualsDigest(v string) bool {
	n, err := strconv.ParseUint(v, 10, 0)
	if err != nil {
		return false
	}

	return n == str.digest
}
