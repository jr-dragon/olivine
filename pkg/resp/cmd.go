package resp

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrProtocol = errors.New("Protocol error")
)

type Command struct {
	raw  Value
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
		cmd:  cmd,
		args: args,
	}, nil
}

func (cmd *Command) Command() string {
	return strings.ToUpper(string(cmd.cmd.data))
}
