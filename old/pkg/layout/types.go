// Package layout implements the tiling engine with container tree management.
package layout

import "github.com/BurntSushi/xgb/xproto"

type ContainerType int

const (
	ContainerLeaf ContainerType = iota
	ContainerVSplit
	ContainerHSplit
)

type Container struct {
	Type     ContainerType
	Parent   *Container
	Children []*Container
	CurFocus int
	Ratio    float64

	Window *ManagedWindow
}

type Rect struct {
	X, Y, W, H int
}

type ManagedWindow struct {
	Frame   xproto.Window
	Client  xproto.Window
	Title   string
	Focused bool
	Geom    Rect
}

func (w *ManagedWindow) SetGeom(r Rect) {
	w.Geom = r
}

type FocusDir int

const (
	FocusDirLeft FocusDir = iota
	FocusDirDown
	FocusDirUp
	FocusDirRight
)
