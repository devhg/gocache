package gfcache

// A ByteView holds an immutable view of bytes.
type ByteView struct {
	//储存真正的缓存值，选择byte类型是为了支持所有的数据类型
	//b 是只读的，使用 ByteSlice() 方法返回一个拷贝，防止缓存值被外部程序修改
	b []byte
}

// Len returns the view's length
func (b ByteView) Len() int {
	return len(b.b)
}

// ByteSlice returns a copy of the data as a byte slice.
func (b ByteView) ByteSlice() []byte {
	return cloneBytes(b.b)
}

// String returns the data as a string, making a copy if necessary.
func (b ByteView) String() string {
	return string(b.b)
}

func cloneBytes(b []byte) []byte {
	bytes := make([]byte, len(b))
	copy(bytes, b)
	return bytes
}
