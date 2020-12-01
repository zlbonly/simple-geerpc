package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)


type GobCodec struct {
	conn io.ReadWriteCloser
	buf *bufio.Writer
	dec *gob.Decoder
	enc *gob.Encoder
}


var _Codec = (*GobCodec)(nil)

/**
1、GobCodec 结构体，这个结构体由四部分构成，
2、conn 是由构建函数传入，通常是通过 TCP 或者 Unix 建立 socket 时得到的链接实例，
3、dec 和 enc 对应 gob 的 Decoder 和 Encoder，
4、buf 是为了防止阻塞而创建的带缓冲的 Writer，一般这么做能提升性能。
 */
func NewGobCodec (conn io.ReadWriteCloser) Codec{
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(buf),
	}
}

func (c *GobCodec) ReadHeader(h *Header) error {
	return c.dec.Decode(h)
}

func (c *GobCodec) ReadBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *GobCodec) Write(h *Header, body interface{}) (err error) {
	defer func() {
		_ = c.buf.Flush()
		if err != nil {
			_ = c.Close()
		}
	}()
	if err = c.enc.Encode(h); err != nil {
		log.Println("rpc: gob error encoding header:", err)
		return
	}
	if err = c.enc.Encode(body); err != nil {
		log.Println("rpc: gob error encoding body:", err)
		return
	}
	return
}

func (c *GobCodec) Close() error {
	return c.conn.Close()
}
