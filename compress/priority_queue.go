package compress

type hp []*Node
func (h hp) Len() int              { return len(h) }
func (h hp) Less(i, j int) bool    { return h[i].freq < h[j].freq }
func (h hp) Swap(i, j int)         { h[i], h[j] = h[j], h[i] }
func (h *hp) Push(v interface{})   { *h = append(*h, v.(*Node)) }
func (h *hp) Pop() (v interface{}) { a := *h; *h, v = a[:len(a)-1], a[len(a)-1]; return }