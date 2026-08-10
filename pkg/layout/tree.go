package layout

func (c *Container) LeafWindows() []*ManagedWindow {
	if c == nil {
		return nil
	}
	if c.Type == ContainerLeaf {
		if c.Window != nil {
			return []*ManagedWindow{c.Window}
		}
		return nil
	}
	var out []*ManagedWindow
	for _, ch := range c.Children {
		out = append(out, ch.LeafWindows()...)
	}
	return out
}

func (c *Container) RemoveFocused() *ManagedWindow {
	if c == nil {
		return nil
	}

	if c.Type == ContainerLeaf {
		return c.Window
	}

	focused := c.Children[c.CurFocus]
	removed := focused.RemoveFocused()
	if removed == nil {
		return nil
	}

	newChildren := make([]*Container, 0, len(c.Children)-1)
	for _, ch := range c.Children {
		if ch != focused {
			newChildren = append(newChildren, ch)
		}
	}
	c.Children = newChildren
	if c.CurFocus >= len(c.Children) {
		c.CurFocus = max(0, len(c.Children)-1)
	}

	return removed
}

func (c *Container) RemoveWindow(w *ManagedWindow) bool {
	if c == nil {
		return false
	}
	if c.Type == ContainerLeaf {
		return c.Window == w
	}
	for i, ch := range c.Children {
		if ch.RemoveWindow(w) {
			newChildren := make([]*Container, 0, len(c.Children)-1)
			for j, c2 := range c.Children {
				if j != i {
					newChildren = append(newChildren, c2)
				}
			}
			c.Children = newChildren
			if c.CurFocus >= len(c.Children) {
				c.CurFocus = max(0, len(c.Children)-1)
			}
			return true
		}
	}
	return false
}

func (c *Container) ResizeFocused(delta float64) {
	if c == nil || c.Type == ContainerLeaf {
		return
	}
	for i, ch := range c.Children {
		if ch.Type == ContainerLeaf && ch.Window != nil && ch.Window.Focused {
			adjustRatio(c.Children, i, delta)
			return
		}
	}
	for _, ch := range c.Children {
		ch.ResizeFocused(delta)
	}
}

func (c *Container) FindParentOf(w *ManagedWindow) *Container {
	if c == nil {
		return nil
	}
	if c.Type == ContainerLeaf {
		return nil
	}
	for _, ch := range c.Children {
		if ch.Type == ContainerLeaf && ch.Window == w {
			return c
		}
		if p := ch.FindParentOf(w); p != nil {
			return p
		}
	}
	return nil
}

func (c *Container) SwapWithNeighbor(swapDir int) bool {
	_ = swapDir
	return false
}

func (c *Container) Count() int {
	if c == nil {
		return 0
	}
	if c.Type == ContainerLeaf {
		return 1
	}
	n := 0
	for _, ch := range c.Children {
		n += ch.Count()
	}
	return n
}

func FindNextLeaf(c *Container) *ManagedWindow {
	if c == nil {
		return nil
	}
	if c.Type == ContainerLeaf {
		return c.Window
	}
	for _, ch := range c.Children {
		if l := FindNextLeaf(ch); l != nil {
			return l
		}
	}
	return nil
}

func FindChildIndex(parent *Container, child *Container) int {
	for i := range parent.Children {
		if parent.Children[i] == child {
			return i
		}
	}
	return -1
}

func adjustRatio(children []*Container, idx int, delta float64) {
	if len(children) <= 1 {
		return
	}
	step := 0.02
	if delta < 0 {
		step = -step
	}
	children[idx].Ratio += step
	if children[idx].Ratio < 0.1 {
		children[idx].Ratio = 0.1
	}
	if children[idx].Ratio > 0.9 {
		children[idx].Ratio = 0.9
	}
	total := 0.0
	for _, ch := range children {
		total += ch.Ratio
	}
	for _, ch := range children {
		ch.Ratio /= total
	}
}
