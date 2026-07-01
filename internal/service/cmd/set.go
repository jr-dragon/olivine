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
	param, err := c.parse(cmd)
	if err != nil {
		return resp.NewSimpleError(err), err
	}

	if err := c.storage.Set(ctx, param); err != nil {
		if errors.Is(err, repo.ErrTypeMismatch) {
			return resp.NewSimpleError(ErrWrongType), err
		}
		if errors.Is(err, repo.ErrCondMismatch) {
			return nil, err
		}
		return nil, fmt.Errorf("%w: %w", ErrStorage, err)
	}

	if param.KeepTTL() && param.cur != nil {
		param.exp = param.cur.ExpiresAt()
	}

	if param.ExpiresAt() != nil &&
		!param.ExpiresAt().IsZero() { // param.exp is set to &time.Time{} as initializing with KEEPTTL is set
		cmd.UpdateAOF(param.expargIdx-1, resp.NewBulkString("PXAT"))
		cmd.UpdateAOF(param.expargIdx, resp.NewBulkString(strconv.FormatInt(param.ExpiresAt().UnixMilli(), 10)))
	}

	if param.GetCurrent() {
		if param.cur == nil {
			return nil, nil
		}
		return resp.NewBulkString(param.cur.String()), nil
	}

	return resp.SimpleString("OK"), nil
}

func (c *Set) parse(cmd *resp.Command) (*setparams, error) {
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

	var (
		expirationOption string
	)

	var p setparams

	state := awaitingKey
	for i, arg := range cmd.Args() {
		switch state {
		case awaitingKey:
			p.key = arg.String()
			state = awaitingValue
		case awaitingValue:
			p.val = arg.String()
			state = awaitingOption
		case awaitingOption:
			switch strings.ToUpper(arg.String()) {
			case "NX":
				p.condType = repo.CondNX
				state = awaitingPostConditionOption
			case "XX":
				p.condType = repo.CondXX
				state = awaitingPostConditionOption
			case "IFEQ":
				p.condType = repo.CondIFEQ
				state = awaitingConditionValue
			case "IFNE":
				p.condType = repo.CondIFNE
				state = awaitingConditionValue
			case "IFDEQ":
				p.condType = repo.CondIFDEQ
				state = awaitingConditionValue
			case "IFDNE":
				p.condType = repo.CondIFDNE
				state = awaitingConditionValue
			case "GET":
				p.get = true
				state = awaitingPostGetOption
			case "EX", "PX", "EXAT", "PXAT":
				p.expargIdx = i + 2
				expirationOption = strings.ToUpper(arg.String())
				state = awaitingExpirationValue
			case "KEEPTTL":
				p.expargIdx = i + 2
				p.keepTTL = true
				p.exp = &time.Time{}
				state = done
			default:
				state = invalid
			}
		case awaitingConditionValue:
			p.condValue = arg.String()
			state = awaitingPostConditionOption
		case awaitingPostConditionOption:
			switch strings.ToUpper(arg.String()) {
			case "GET":
				p.get = true
				state = awaitingPostGetOption
			case "EX", "PX", "EXAT", "PXAT":
				p.expargIdx = i + 2
				expirationOption = strings.ToUpper(arg.String())
				state = awaitingExpirationValue
			case "KEEPTTL":
				p.expargIdx = i + 2
				p.keepTTL = true
				p.exp = &time.Time{}
				state = done
			default:
				state = invalid
			}
		case awaitingPostGetOption:
			switch strings.ToUpper(arg.String()) {
			case "EX", "PX", "EXAT", "PXAT":
				p.expargIdx = i + 2
				expirationOption = strings.ToUpper(arg.String())
				state = awaitingExpirationValue
			case "KEEPTTL":
				p.expargIdx = i + 2
				p.keepTTL = true
				p.exp = &time.Time{}
				state = done
			default:
				state = invalid
			}
		case awaitingExpirationValue:
			exp, err := parseExpiration(expirationOption, arg.String())
			if err != nil {
				return nil, fmt.Errorf("%w: %w", ErrSyntax, err)
			}
			p.exp = &exp
			state = done
		case done:
			state = invalid
		}
	}

	if state != awaitingOption && state != awaitingPostConditionOption && state != awaitingPostGetOption && state != done {
		return nil, fmt.Errorf("%w: invalid SET syntax", ErrSyntax)
	}

	return &p, nil
}

func parseExpiration(option, value string) (time.Time, error) {
	switch option {
	case "EX", "PX":
		unit := "s"
		if option == "PX" {
			unit = "ms"
		}

		duration, err := parseDuration(value, unit)
		if err != nil {
			return time.Time{}, err
		}

		return time.Now().Add(duration), nil
	case "EXAT", "PXAT":
		timestamp, err := strconv.Atoi(value)
		if err != nil {
			return time.Time{}, err
		}
		if timestamp <= 0 {
			return time.Time{}, errors.New("invalid expire time")
		}
		if option == "EXAT" {
			return time.Unix(int64(timestamp), 0), nil
		}

		return time.UnixMilli(int64(timestamp)), nil
	}

	return time.Time{}, fmt.Errorf("unknown expiration option %q", option)
}

func parseDuration(s, unit string) (time.Duration, error) {
	if _, err := strconv.Atoi(s); err != nil {
		return 0, fmt.Errorf("invalid expire time: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(s)
	sb.WriteString(unit)

	if d, err := time.ParseDuration(sb.String()); err != nil {
		return d, err
	} else if d <= 0 {
		return d, errors.New("invalid expire time")
	} else {
		return d, nil
	}
}

type setparams struct {
	key string
	val string

	// NX|XX|IFEQ|IFNE|IFDEQ|IFDNE
	condType  repo.Cond
	condValue string

	// GET
	get bool
	cur *object.String

	// EX|PX|EXAT|PXAT
	exp     *time.Time
	keepTTL bool

	expargIdx int
}

var _ repo.SetStringParam = &setparams{}

func (p *setparams) Obj() object.Object {
	return object.NewString(p.key, p.val, p.exp)
}

func (p *setparams) CondType() repo.Cond {
	return p.condType
}

func (p *setparams) CondValue() string {
	return p.condValue
}

func (p *setparams) ExpiresAt() *time.Time {
	return p.exp
}

func (p *setparams) KeepTTL() bool {
	return p.keepTTL
}

func (p *setparams) GetCurrent() bool {
	return p.get
}

func (p *setparams) SetCurrent(cur *object.String) {
	p.cur = cur
}
