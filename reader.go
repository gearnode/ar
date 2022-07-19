// Copyright (c) 2022 Bryan Frimin <bryan@frimin.fr>.
//
// Permission to use, copy, modify, and/or distribute this software for
// any purpose with or without fee is hereby granted, provided that the
// above copyright notice and this permission notice appear in all
// copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL
// WARRANTIES WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE
// AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL
// DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR
// PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER
// TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package ar

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"time"
)

// Reader provides sequential access to the contents of a ar archive.
//
// Reader.Next advances to the next file in the archive (including the
// first), and then Reader can be treated as an io.Reader to access the
// file's data.
type Reader struct {
	io io.Reader
	n  int64
	p  int64
}

// NewReader creates a new Reader reading from r.
//
// An error is returned when the reader is not an ar file.
func NewReader(r io.Reader) (*Reader, error) {
	mb := make([]byte, 8)
	if _, err := io.ReadFull(r, mb); err != nil {
		return nil, fmt.Errorf("cannot read magic byte: %w", err)
	}

	if !bytes.Equal(mb, []byte("!<arch>\n")) {
		return nil, fmt.Errorf("invalid magic bytes, expected"+
			" %q got %q", MagicString, string(mb))
	}

	return &Reader{io: r}, nil
}

// Next advances to the next entry in the ar archive.
//
// The Header.Size determines how many bytes can be read for the next
// file. Any remaining data in the current file is automatically
// discarded.
//
// io.EOF is returned at the end of the input.
func (r *Reader) Next() (*Header, error) {
	err := r.skipUnread()
	if err != nil {
		return nil, err
	}

	buf := make([]byte, HeaderByteSize)
	size, err := io.ReadFull(r.io, buf)
	if err != nil {
		return nil, err
	}

	if size != HeaderByteSize {
		return nil, fmt.Errorf("invalid file header size, "+
			"expected %d got %d", HeaderByteSize, size)
	}

	if buf[58] != '`' && buf[59] != '\n' {
		return nil, fmt.Errorf("invalid header ending bytes")
	}

	date, err := readInt(buf[16:28])
	if err != nil {
		return nil, fmt.Errorf("cannot parse file timestamp: %w",
			err)
	}

	uid, err := readInt(buf[28:34])
	if err != nil {
		return nil, fmt.Errorf("cannot parse file owner id: %w",
			err)
	}

	gid, err := readInt(buf[34:40])
	if err != nil {
		return nil, fmt.Errorf("cannot parse file group id: %w",
			err)
	}

	mode, err := readOctal(buf[40:48])
	if err != nil {
		return nil, fmt.Errorf("cannot parse file mode: %w", err)
	}

	fsize, err := readInt(buf[48:58])
	if err != nil {
		return nil, fmt.Errorf("cannot parse file size: %w", err)
	}

	header := Header{
		Name: readString(buf[0:16]),
		Date: time.Unix(date, 0),
		Uid:  uid,
		Gid:  gid,
		Mode: mode,
		Size: fsize,
	}

	r.n = header.Size
	r.p = header.Size % 2

	return &header, nil
}

// Read reads from the current file in the ar archive.
//
// It returns (0, io.EOF) when it reaches the end of that file, until
// Next is called to advance to the next file.
func (r *Reader) Read(b []byte) (int, error) {
	if r.n == 0 {
		return 0, io.EOF
	}

	if int64(len(b)) > r.n {
		b = b[0:r.n]
	}

	n, err := r.io.Read(b)
	r.n -= int64(n)

	return n, err
}

func (r *Reader) skipUnread() error {
	s := r.n + r.p
	if _, err := io.CopyN(ioutil.Discard, r.io, s); err != nil {
		return err
	}

	r.n = 0
	r.p = 0

	return nil

}

func readString(b []byte) string {
	i := len(b) - 1
	for i > 0 && b[i] == ' ' {
		i--
	}
	return string(b[0 : i+1])
}

func readInt(b []byte) (int64, error) {
	i := len(b) - 1
	for i > 0 && b[i] == ' ' {
		i--
	}

	n, err := strconv.ParseInt(string(b[0:i+1]), 10, 64)
	return n, err
}

func readOctal(b []byte) (int64, error) {
	i := len(b) - 1
	for i > 0 && b[i] == ' ' {
		i--
	}

	n, err := strconv.ParseInt(string(b[0:i+1]), 8, 64)
	return n, err
}
