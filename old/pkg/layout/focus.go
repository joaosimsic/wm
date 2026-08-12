package layout

func DirDist(dir FocusDir, ref, r Rect) (int, bool) {
	switch dir {
	case FocusDirLeft:
		if r.X+r.W <= ref.X {
			return ref.X - (r.X + r.W) + abs((r.Y+r.H/2)-(ref.Y+ref.H/2)), true
		}
	case FocusDirDown:
		if r.Y >= ref.Y+ref.H {
			return abs((r.X+r.W/2)-(ref.X+ref.W/2)) + r.Y - (ref.Y + ref.H), true
		}
	case FocusDirUp:
		if r.Y+r.H <= ref.Y {
			return abs((r.X+r.W/2)-(ref.X+ref.W/2)) + ref.Y - (r.Y + r.H), true
		}
	case FocusDirRight:
		if r.X >= ref.X+ref.W {
			return abs((r.Y+r.H/2)-(ref.Y+ref.H/2)) + r.X - (ref.X + ref.W), true
		}
	}
	return 0, false
}

func (c *Container) FocusInDir(dir FocusDir) *ManagedWindow {
	windows := c.LeafWindows()
	if len(windows) == 0 {
		return nil
	}

	var focused *ManagedWindow
	fIdx := -1
	for i, w := range windows {
		if w.Focused {
			focused = w
			fIdx = i
			break
		}
	}
	if focused == nil {
		return windows[0]
	}

	rect := focused.Geom
	var best *ManagedWindow
	bestDist := int(1e9)

	for i, w := range windows {
		if i == fIdx {
			continue
		}
		if d, ok := DirDist(dir, rect, w.Geom); ok && d < bestDist {
			bestDist = d
			best = w
		}
	}

	if best != nil {
		return best
	}
	return focused
}

func (c *Container) FindFocused() *Container {
	if c == nil {
		return nil
	}
	if c.Type == ContainerLeaf {
		return c
	}
	if c.CurFocus < len(c.Children) {
		return c.Children[c.CurFocus].FindFocused()
	}
	return nil
}

func (c *Container) FocusNext() *ManagedWindow {
	windows := c.LeafWindows()
	if len(windows) == 0 {
		return nil
	}
	for i, w := range windows {
		if w.Focused {
			next := (i + 1) % len(windows)
			return windows[next]
		}
	}
	return windows[0]
}

func (c *Container) FocusPrev() *ManagedWindow {
	windows := c.LeafWindows()
	if len(windows) == 0 {
		return nil
	}
	for i, w := range windows {
		if w.Focused {
			prev := i - 1
			if prev < 0 {
				prev = len(windows) - 1
			}
			return windows[prev]
		}
	}
	return windows[0]
}

func (c *Container) MoveFocus(dir string) {
	if c == nil || c.Type == ContainerLeaf || len(c.Children) == 0 {
		return
	}
	switch dir {
	case "l":
		c.CurFocus = max(0, c.CurFocus-1)
	case "r":
		c.CurFocus = min(len(c.Children)-1, c.CurFocus+1)
	}
}

func (c *Container) SetFocusTo(w *ManagedWindow) bool {
	if c == nil {
		return false
	}
	if c.Type == ContainerLeaf {
		return c.Window == w
	}
	for i, ch := range c.Children {
		if ch.SetFocusTo(w) {
			c.CurFocus = i
			return true
		}
	}
	return false
}
