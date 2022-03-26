package mycache

type ByteView struct {
	b []byte
}

func (b ByteView) Len() int {
	return len(b.b)
}

func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

func cloneBytes(src []byte) []byte {
	b := make([]byte, len(src))
	copy(b, src)
	return b
}

func (v ByteView) String() string {
	return string(v.b)
}
