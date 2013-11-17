// Simple bytepool for Nabu
package autocomplete

// A pool of fixed-length byte buffers
type Pool struct {
  buffers chan*Buffer
}

// Create a new byte pool holding count item, each item of size bytes
func newPool(count, size int) *Pool {
  p := &Pool{
    buffers: make(chan*Buffer, count),
  }
  for i := 0; i < count; i++ {
    p.buffers <- &Buffer{
      pool: p,
      bytes: make([]byte, size),
    }
  }
  return p
}

// Get a buffer
func (p *Pool) Checkout() *Buffer {
  return <- p.buffers
}

// A fixed-lenght byte buffer
type Buffer struct {
  pool *Pool
  length int
  bytes []byte
}

// Writes a byte into the vuffer
func (b *Buffer) Write(bt byte) int {
  b.bytes[b.length] = bt
  b.length++
  return b.length
}

// Get the buffer's bytes
func (b *Buffer) Bytes() []byte {
  return b.bytes[0:b.length]
}

// Release the buffer back to the pool
func (b *Buffer) Close() {
  b.length = 0
  b.pool.buffers <- b
}
