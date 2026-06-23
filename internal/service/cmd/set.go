package cmd

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"olivine/internal/repo"
	"olivine/internal/repo/object"
	"olivine/pkg/resp"
)

type Set struct {
	storage repo.Storage
}

func NewSet(storage repo.Storage) *Set {
	return &Set{storage: storage}
}

func (c *Set) Command() string {
	return "SET"
}

func (c *Set) Exec(ctx context.Context, cmd *resp.Command) (resp.Value, error) {
	parsed, err := c.parse(cmd)
	if err != nil {
		return nil, err
	}

	if err := c.storage.Set(ctx, parsed); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStorage, err)
	}

	return resp.SimpleString("OK"), nil
}

func (c *Set) parse(cmd *resp.Command) (p parsedSet, err error) {
	const (
		awaitingKey = iota
		awaitingValue
		awaitingOption
		awaitingConditionValue
		awaitingPostConditionOption
		awaitingPostGetOption
		awaitingExpirationValue
		done
		invalid
	)

	state := awaitingKey
	var conditionType int
	var expirationOption string

	for _, arg := range cmd.Args() {
		switch state {
		case awaitingKey:
			p.K = arg.String()
			state = awaitingValue
		case awaitingValue:
			p.V = arg.String()
			state = awaitingOption
		case awaitingOption:
			switch strings.ToUpper(arg.String()) {
			case "NX":
				p.Cond.Typ = nx
				state = awaitingPostConditionOption
			case "XX":
				p.Cond.Typ = xx
				state = awaitingPostConditionOption
			case "IFEQ":
				conditionType = ifeq
				state = awaitingConditionValue
			case "IFNE":
				conditionType = ifne
				state = awaitingConditionValue
			case "IFDEQ":
				conditionType = ifdeq
				state = awaitingConditionValue
			case "IFDNE":
				conditionType = ifdne
				state = awaitingConditionValue
			case "GET":
				p.Get = true
				state = awaitingPostGetOption
			case "EX", "PX", "EXAT", "PXAT":
				expirationOption = strings.ToUpper(arg.String())
				state = awaitingExpirationValue
			case "KEEPTTL":
				p.Exp = new(time.Time)
				state = done
			default:
				state = invalid
			}
		case awaitingConditionValue:
			p.Cond = parseSetCond{Typ: conditionType, Val: arg.String()}
			state = awaitingPostConditionOption
		case awaitingPostConditionOption:
			switch strings.ToUpper(arg.String()) {
			case "GET":
				p.Get = true
				state = awaitingPostGetOption
			case "EX", "PX", "EXAT", "PXAT":
				expirationOption = strings.ToUpper(arg.String())
				state = awaitingExpirationValue
			case "KEEPTTL":
				p.Exp = new(time.Time)
				state = done
			default:
				state = invalid
			}
		case awaitingPostGetOption:
			switch strings.ToUpper(arg.String()) {
			case "EX", "PX", "EXAT", "PXAT":
				expirationOption = strings.ToUpper(arg.String())
				state = awaitingExpirationValue
			case "KEEPTTL":
				p.Exp = new(time.Time)
				state = done
			default:
				state = invalid
			}
		case awaitingExpirationValue:
			p.Exp, err = parseExpiration(expirationOption, arg.String())
			if err != nil {
				return parsedSet{}, fmt.Errorf("%w: %w", ErrSyntax, err)
			}
			state = done
		case done:
			state = invalid
		}
	}

	if state != awaitingOption && state != awaitingPostConditionOption && state != awaitingPostGetOption && state != done {
		return parsedSet{}, fmt.Errorf("%w: invalid SET syntax", ErrSyntax)
	}

	return p, nil
}

func parseExpiration(option, value string) (*time.Time, error) {
	switch option {
	case "EX", "PX":
		unit := "s"
		if option == "PX" {
			unit = "ms"
		}

		duration, err := parseDuration(value, unit)
		if err != nil {
			return nil, err
		}

		return new(time.Now().Add(duration)), nil
	case "EXAT", "PXAT":
		timestamp, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		if timestamp < 0 {
			return nil, errors.New("invalid expire time")
		}
		if option == "EXAT" {
			return new(time.Unix(int64(timestamp), 0)), nil
		}

		return new(time.UnixMilli(int64(timestamp))), nil
	}

	return nil, fmt.Errorf("unknown expiration option %q", option)
}

func parseDuration(s, unit string) (time.Duration, error) {
	var sb strings.Builder
	sb.WriteString(s)
	sb.WriteString(unit)

	if d, err := time.ParseDuration(sb.String()); err != nil {
		return d, err
	} else if d < 0 {
		return d, errors.New("invalid expire time")
	} else {
		return d, nil
	}
}

const (
	nx    = iota + 1 // set value only not exists
	xx               // set value only exists
	ifeq             // set only value == cond.Val
	ifne             // set only value != cond.Val
	ifdeq            // set only XXH3(value) == cond.Val
	ifdne            // setonly XXH3(value) != cond.Val
)

type parseSetCond struct {
	Typ int
	Val string
}

type parsedSet struct {
	K    string
	V    string
	Cond parseSetCond
	Get  bool
	Exp  *time.Time
}

var _ repo.SetParam = parsedSet{}

func (parsed parsedSet) Obj() object.Object {
	return object.NewString(parsed.K, parsed.V, parsed.Exp)
}
