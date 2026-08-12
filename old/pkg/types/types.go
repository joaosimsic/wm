// Package types provides shared mode, action, keybind, and command types.
package types

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/xgb/xproto"
)

type Mode int

const (
	ModeNormal Mode = iota
	ModeCommand
)

func (m Mode) String() string {
	switch m {
	case ModeNormal:
		return "NORMAL"
	case ModeCommand:
		return "COMMAND"
	}
	return "???"
}

type Action int

const (
	ActionFocusLeft Action = iota
	ActionFocusDown
	ActionFocusUp
	ActionFocusRight
	ActionMoveLeft
	ActionMoveDown
	ActionMoveUp
	ActionMoveRight
	ActionResizeLeft
	ActionResizeDown
	ActionResizeUp
	ActionResizeRight
	ActionCmdMode
	ActionClose
	ActionTerminal
	ActionWS1
	ActionWS2
	ActionWS3
	ActionWS4
	ActionWS5
	ActionWS6
	ActionWS7
	ActionWS8
	ActionWS9
	ActionWS10
	ActionWSNext
	ActionWSPrev
)

var (
	Mod4Mask    uint16 = xproto.KeyButMaskMod4
	ShiftMask   uint16 = xproto.KeyButMaskShift
	ControlMask uint16 = xproto.KeyButMaskControl
	LockMask    uint16 = xproto.KeyButMaskLock
	Mod2Mask    uint16 = 16
	Mod5Mask    uint16 = 128
	IgnoredMods uint16 = LockMask | Mod2Mask | Mod5Mask
)

type KeyBind struct {
	Mods    uint16
	Keycode xproto.Keycode
	Action  Action
}

type Cmd struct {
	Name string
	Args string
}

func ParseCommand(input string) (*Cmd, error) {
	if !strings.HasPrefix(input, ":") {
		return nil, fmt.Errorf("command must start with ':'")
	}
	body := strings.TrimPrefix(input, ":")
	body = strings.TrimSpace(body)

	parts := strings.SplitN(body, " ", 2)
	name := parts[0]
	args := ""
	if len(parts) > 1 {
		args = strings.TrimSpace(parts[1])
	}

	return &Cmd{Name: name, Args: args}, nil
}
