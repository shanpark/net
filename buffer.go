package net

// A Buffer is a variable-sized buffer of bytes with Read, Write, Commit, Rollback methods.
type Buffer struct {
	buf []byte
	si  int
	ri  int
	wi  int
}

// NewBuffer returns a buffer.
func NewBuffer(size int) *Buffer {
	buffer := new(Buffer)
	buffer.buf = make([]byte, size)
	buffer.si = 0
	buffer.ri = 0
	buffer.wi = 0
	return buffer
}

// Read implements io.Reader interface.
func (b *Buffer) Read(p []byte) (n int, err error) {
	n = copy(p, b.buf[b.ri:b.wi])
	b.ri += n
	return n, nil
}

// Write implements io.Writer interface.
func (b *Buffer) Write(p []byte) (n int, err error) {
	b.Reserve(len(p))
	n = copy(b.buf[b.wi:], p)
	b.wi += n
	return n, nil
}

// Readable returns the number of bytes that can be read.
func (b *Buffer) Readable() int {
	return b.wi - b.ri
}

// Data returns the byte slice of data that can be read.
// No allocation occurs. And this doesn't consume readable data from the buffer.
// The returned data is still available from the buffer.
func (b *Buffer) Data() []byte {
	return b.buf[b.ri:b.wi]
}

// DataConsume consumes 'n' bytes from the buffer.
func (b *Buffer) DataConsume(n int) {
	b.ri += n
	if b.ri > b.wi {
		panic("over consumed")
	}
}

// Flush clears all data from the buffer.
func (b *Buffer) Flush() {
	b.si = 0
	b.ri = 0
	b.wi = 0
}

// Reserve reserves the writable space in the buffer.
func (b *Buffer) Reserve(need int) {
	space := len(b.buf) - b.wi
	if need > space {
		b.grow(need - space)
	}
}

// Buffer returns the byte slice of the writable space from the buffer.
// No allocation occurs. And this doesn't consume writable space from the buffer.
// If you writes some data to the slices, you should call BufferConsume() methods.
func (b *Buffer) Buffer() []byte {
	return b.buf[b.wi:]
}

// BufferConsume consumes 'n' bytes of the writable space from the buffer
func (b *Buffer) BufferConsume(n int) {
	b.wi += n
	if b.wi > len(b.buf) {
		panic("buffer overflow")
	}
}

// Commit applies the state of the buffer changed by Read().
func (b *Buffer) Commit() {
	b.si = b.ri
}

// Rollback discards the state of the buffer changed by Read() as if you hadn't read it.
func (b *Buffer) Rollback() {
	b.ri = b.si
}

func (b *Buffer) grow(need int) {
	space := len(b.buf) - b.wi + b.si
	if space > need {
		b.wi = copy(b.buf, b.buf[b.si:b.wi])
		b.ri = b.ri - b.si
		b.si = 0
	} else {
		buf := make([]byte, len(b.buf)*2)
		b.wi = copy(buf, b.buf[b.si:b.wi])
		b.ri = b.ri - b.si
		b.si = 0
		b.buf = buf
	}
}
