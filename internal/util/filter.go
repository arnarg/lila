package util

import (
	"bufio"
	"bytes"
	"io"
)

type Filter struct {
	reader *bufio.Reader
	prefix []byte
	plen   int
	unread []byte
	eof    bool
}

func NewFilter(r io.Reader, prefix string) io.Reader {
	return &Filter{
		reader: bufio.NewReader(r),
		prefix: []byte(prefix),
		plen:   len(prefix),
	}
}

func (r *Filter) Read(p []byte) (n int, err error) {
	for {
		// Write unread data from previous read.
		if len(r.unread) > 0 {
			m := copy(p[n:], r.unread)
			n += m
			r.unread = r.unread[m:]
			if len(r.unread) > 0 {
				return n, nil
			}
		}

		// The underlying Reader already returned EOF, do not read again.
		if r.eof {
			return n, io.EOF
		}

		// Read new line, including delim.
		line, err := r.reader.ReadBytes('\n')

		if err == io.EOF {
			r.eof = true
		}

		// No new data, do not block.
		if len(line) == 0 {
			return n, err
		}

		if bytes.HasPrefix(line, r.prefix) {
			r.unread = line[r.plen:]
		}

		if err != nil {
			if err == io.EOF && len(r.unread) > 0 {
				// The underlying Reader already returned EOF, but we still
				// have some unread data to send, thus clear the error.
				return n, nil
			}
			return n, err
		}
	}
}
