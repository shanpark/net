package net

// A Buffer is a variable-sized buffer of bytes with Read, Write, Commit, Rollback methods.
type Buffer struct {
	buf []byte
	si  int
	ri  int
	wi  int
}

// NewBuffer returns a buffer.
func NewBuffer() *Buffer {
	buffer := new(Buffer)
	buffer.buf = make([]byte, 4096)
	buffer.si = 0
	buffer.ri = 0
	buffer.wi = 0
	return buffer
}

func (b *Buffer) Read(p []byte) (n int, err error) {
	n = copy(p, b.buf[b.ri:b.wi])
	b.ri += n
	return n, nil
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	b.Reserve(len(p))
	n = copy(b.buf[b.wi:], p)
	b.wi += n
	return n, nil
}

func (b *Buffer) Readable() int {
	return b.wi - b.ri
}

func (b *Buffer) Data() []byte {
	return b.buf[b.ri:b.wi]
}

func (b *Buffer) DataConsume(n int) {
	b.ri += n
	if b.ri > b.wi {
		panic("over consumed")
	}
}

func (b *Buffer) Flush() {
	b.si = 0
	b.ri = 0
	b.wi = 0
}

func (b *Buffer) Reserve(need int) {
	space := len(b.buf) - b.wi
	if need > space {
		b.grow(need - space)
	}
}

func (b *Buffer) Buffer() []byte {
	return b.buf[b.wi:]
}

func (b *Buffer) BufferConsume(n int) {
	b.wi += n
	if b.wi > len(b.buf) {
		panic("buffer overflow")
	}
}

// Commit applies the state of the buffer changed by Read().
func (b *Buffer) commit() {
	b.si = b.ri
}

// Rollback discards the state of the buffer changed by Read() as if you hadn't read it.
func (b *Buffer) rollback() {
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
