package resp

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

var (
	ErrProtocol = errors.New("Protocol error")
)

var (
	dirtyCommands = map[string]struct{}{
		"SET": {},
	}
)

type Command struct {
	raw []BulkString
	aof []BulkString
}

func ReadCommand(rd *Reader) (*Command, error) {
	values, err := rd.ReadCommand()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrProtocol, err)
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("%w: empty array", ErrProtocol)
	}

	return &Command{
		raw: values,
	}, nil
}

func NewTestCommand(v Array) *Command {
	strs := make([]BulkString, 0, len(v.data))
	for _, s := range v.data {
		strs = append(strs, s.(BulkString))
	}

	return &Command{
		raw: strs,
		aof: slices.Clone(strs),
	}
}

func (cmd *Command) Command() string {
	return strings.ToUpper(string(cmd.raw[0].data))
}

func (cmd *Command) Args() []BulkString {
	return cmd.raw[1:]
}

func (cmd *Command) Dirty() bool {
	_, isdirty := dirtyCommands[cmd.Command()]
	return isdirty
}

func (cmd *Command) UpdateAOF(i int, v BulkString) {
	if cmd.aof == nil {
		cmd.aof = slices.Clone(cmd.raw)
	}

	if i < len(cmd.aof) {
		cmd.aof[i] = v
	} else {
		cmd.aof = append(cmd.aof, v)
	}
}

func (cmd *Command) MarshalAOF() []byte {
	if cmd.aof == nil {
		return marshalCommand(cmd.raw)
	}
	return marshalCommand(cmd.aof)
}

func (cmd *Command) Marshal() []byte {
	return marshalCommand(cmd.raw)
}

func marshalCommand(values []BulkString) []byte {
	var buf bytes.Buffer

	buf.WriteByte(MAGIC_ARRAY)
	buf.WriteString(strconv.Itoa(len(values)))
	buf.WriteString(SENTINEL)
	for _, value := range values {
		buf.Write(value.Marshal())
	}

	return buf.Bytes()
}
