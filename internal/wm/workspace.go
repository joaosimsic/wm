package wm

type Workspace struct {
	id   int
	list []*Client
}

func (w *Workspace) add(c *Client) {
	w.list = append(w.list, c)
}

func (w *Workspace) remove(c *Client) {
	i, ok := w.index(c)
	if !ok {
		return
	}

	w.list = append(w.list[:i], w.list[i+1:]...)
}

func (w *Workspace) index(c *Client) (int, bool) {
	for i, client := range w.list {
		if client == c {
			return i, true
		}
	}

	return 0, false
}
