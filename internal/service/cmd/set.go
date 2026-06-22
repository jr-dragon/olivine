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
	args := cmd.Args()
	if len(args) < 2 {
		return nil, fmt.Errorf("%w: argument count mismatch: expect '%d' got '%d'", ErrValidation, len(args), 2)
	}

	k, v := args[0], args[1]
	exp, err := parse(args[2:])
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrValidation, err)
	}

	if err := c.storage.Set(ctx, object.NewString(k.String(), v.String(), exp)); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStorage, err)
	}

	return resp.SimpleString("OK"), nil
}

// TODO: SET {key} {value} [NX|XX|IFEQ {ifeq-val}|IFNE {ifne-val}|IFDEQ {ifdeq-value}|IFDNE {ifdne-value} [GET] [EX {sec}|PX {msec}|EXAT {uxtime}|PXAT {uxptime}|KEEPTTL]
func parse(args []resp.BulkString) (_exp *time.Time, err error) {
	if len(args) == 0 {
		return nil, nil
	}

	// Note: it is not a valid parser, for parsing EXPIRE/TTL feature only
	for i := 0; i < len(args)-1; i++ {
		arg := args[i]
		switch strings.ToUpper(arg.String()) {
		case "EX":
			duration, err := parseDuration(args[i+1].String(), "s")
			if err != nil {
				return nil, err
			}

			return new(time.Now().Add(duration)), nil
		case "PX":
			duration, err := parseDuration(args[i+1].String(), "ms")
			if err != nil {
				return nil, err
			}

			return new(time.Now().Add(duration)), nil
		case "EXAT":
			usec, err := strconv.Atoi(args[i+1].String())
			if err != nil {
				return nil, err
			}
			if usec < 0 {
				return nil, errors.New("invalid expire time")
			}

			return new(time.Unix(int64(usec), 0)), nil
		case "PXAT":
			umsec, err := strconv.Atoi(args[i+1].String())
			if err != nil {
				return nil, err
			}
			if umsec < 0 {
				return nil, errors.New("invalid expire time")
			}

			return new(time.UnixMilli(int64(umsec))), nil
		}
	}

	return nil, nil
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
