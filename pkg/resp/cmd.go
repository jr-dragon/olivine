package resp

import (
	"errors"
	"fmt"
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
	raw  Value
	aof  Array
	cmd  BulkString
	args []BulkString
}

func ReadCommand(rd *Reader) (*Command, error) {
	v, err := rd.Read()
	if err != nil {
		return nil, err
	}

	arr, ok := v.(Array)
	if !ok {
		return nil, fmt.Errorf("%w: expected array, got %T(%+v)", ErrProtocol, v, v)
	}
	if arr.null || len(arr.data) == 0 {
		return nil, fmt.Errorf("%w: empty array", ErrProtocol)
	}

	cmd, ok := arr.data[0].(BulkString)
	if !ok {
		return nil, fmt.Errorf("%w: command expected bulk string, got %T(%+v)", ErrProtocol, arr.data[0], arr.data[0])
	}

	args := make([]BulkString, 0, len(arr.data)-1)
	for i := range arr.data[1:] {
		arg, ok := arr.data[i+1].(BulkString)
		if !ok {
			return nil, fmt.Errorf("%w: argument [%d] expected bulk string, got %T(%+v)", ErrProtocol, i, arr.data[i+1], arr.data[i+1])
		}
		args = append(args, arg)
	}

	return &Command{
		raw:  v,
		aof:  arr.Clone(),
		cmd:  cmd,
		args: args,
	}, nil
}

func NewTestCommand(v Array) *Command {
	strs := make([]BulkString, 0, len(v.data))
	for _, s := range v.data {
		strs = append(strs, s.(BulkString))
	}

	var args []BulkString
	if len(strs) == 1 {
		args = make([]BulkString, 0)
	} else {
		args = strs[1:]
	}

	return &Command{
		raw:  v,
		aof:  v.Clone(),
		cmd:  strs[0],
		args: args,
	}
}

func (cmd *Command) Command() string {
	return strings.ToUpper(string(cmd.cmd.data))
}

func (cmd *Command) Args() []BulkString {
	return cmd.args
}

func (cmd *Command) Dirty() bool {
	_, isdirty := dirtyCommands[cmd.Command()]
	return isdirty
}

func (cmd *Command) UpdateAOF(i int, v Value) {
	if i < len(cmd.aof.data) {
		cmd.aof.data[i] = v
	} else {
		cmd.aof.data = append(cmd.aof.data, v)
	}
}

func (cmd *Command) MarshalAOF() []byte {
	return cmd.aof.Marshal()
}

func (cmd *Command) Marshal() []byte {
	return cmd.raw.Marshal()
}
