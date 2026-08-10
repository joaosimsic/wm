package layout

func (c *Container) SetRatios() {
	if c.Type == ContainerLeaf || len(c.Children) == 0 {
		return
	}
	n := len(c.Children)
	for i := range c.Children {
		if c.Children[i].Ratio == 0 {
			c.Children[i].Ratio = 1.0 / float64(n)
		}
	}
}

func (c *Container) FixChildren() {
	for _, ch := range c.Children {
		ch.Parent = c
	}
}

func (c *Container) Tile(area Rect, bw int, gap int) {
	if c == nil {
		return
	}
	if c.Type == ContainerLeaf {
		if c.Window != nil {
			c.Window.SetGeom(area)
		}
		return
	}

	c.SetRatios()

	totalRatio := 0.0
	for _, ch := range c.Children {
		totalRatio += ch.Ratio
	}

	var rects []Rect
	switch c.Type {
	case ContainerVSplit:
		usableW := area.W - gap*(len(c.Children)-1) - bw*2*len(c.Children)
		x := area.X + bw
		for _, ch := range c.Children {
			w := max(int(float64(usableW)*ch.Ratio/totalRatio), 50)
			rects = append(rects, Rect{
				X: x, Y: area.Y + bw,
				W: w, H: area.H - bw*2,
			})
			x += w + gap + bw*2
		}
	case ContainerHSplit:
		usableH := area.H - gap*(len(c.Children)-1) - bw*2*len(c.Children)
		y := area.Y + bw
		for _, ch := range c.Children {
			h := max(int(float64(usableH)*ch.Ratio/totalRatio), 50)
			rects = append(rects, Rect{
				X: area.X + bw, Y: y,
				W: area.W - bw*2, H: h,
			})
			y += h + gap + bw*2
		}
	}

	for i, ch := range c.Children {
		area := rects[i]
		if ch.Type != ContainerLeaf {
			ch.Tile(area, bw, gap)
			continue
		}

		if ch.Window == nil {
			continue
		}

		ch.Window.SetGeom(area)
	}
}
