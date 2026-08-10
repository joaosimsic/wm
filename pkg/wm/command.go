package wm

import (
	"fmt"
	"strconv"

	"wm/pkg/layout"
	"wm/pkg/types"
)

var commandTable = map[string]func(*WM, *types.Cmd) error{
	"sp":     cmdSplit,
	"vsp":    cmdVSplit,
	"q":      cmdClose,
	"q!":     cmdKill,
	"w":      cmdWorkspace,
	"tabnew": cmdNewWorkspace,
	"tabn":   cmdTabNext,
	"tabp":   cmdTabPrev,
	"exec":   cmdExec,
	"reload": cmdReload,
}

func (wm *WM) executeCommand(cmd *types.Cmd) error {
	fn, ok := commandTable[cmd.Name]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmd.Name)
	}
	return fn(wm, cmd)
}

func cmdSplit(wm *WM, cmd *types.Cmd) error {
	prog := cmd.Args
	if prog == "" {
		prog = wm.conf.Terminal
	}
	wm.pendingSplit = true
	wm.pendingDir = layout.ContainerHSplit
	return wm.launch(prog)
}

func cmdVSplit(wm *WM, cmd *types.Cmd) error {
	prog := cmd.Args
	if prog == "" {
		prog = wm.conf.Terminal
	}
	wm.pendingSplit = true
	wm.pendingDir = layout.ContainerVSplit
	return wm.launch(prog)
}

func cmdClose(wm *WM, _ *types.Cmd) error {
	wm.closeFocused()
	return nil
}

func cmdKill(wm *WM, _ *types.Cmd) error {
	if wm.focused == nil {
		return nil
	}
	return wm.killWindow(wm.focused.Client)
}

func cmdWorkspace(wm *WM, cmd *types.Cmd) error {
	if cmd.Args == "" {
		return nil
	}
	n, err := strconv.Atoi(cmd.Args)
	if err != nil {
		return fmt.Errorf("invalid workspace number: %s", cmd.Args)
	}
	n--
	if n < 0 || n >= len(wm.workspaces) {
		return fmt.Errorf("workspace %d out of range", n+1)
	}
	wm.switchWorkspace(n)
	return nil
}

func cmdNewWorkspace(wm *WM, _ *types.Cmd) error {
	for i, ws := range wm.workspaces {
		if ws.Root == nil {
			wm.switchWorkspace(i)
			return nil
		}
	}
	return fmt.Errorf("all workspaces in use")
}

func cmdTabNext(wm *WM, _ *types.Cmd) error {
	wm.switchWorkspace((wm.currentWS + 1) % len(wm.workspaces))
	return nil
}

func cmdTabPrev(wm *WM, _ *types.Cmd) error {
	prev := wm.currentWS - 1
	if prev < 0 {
		prev = len(wm.workspaces) - 1
	}
	wm.switchWorkspace(prev)
	return nil
}

func cmdExec(wm *WM, cmd *types.Cmd) error {
	if cmd.Args == "" {
		return fmt.Errorf(":exec requires a command")
	}
	return wm.launch(cmd.Args)
}

func cmdReload(wm *WM, _ *types.Cmd) error {
	return wm.reloadConfig()
}
