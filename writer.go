// Copyright (c) 2022 Bryan Frimin <bryan@frimin.fr>.>
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
	"errors"
	"io"
	"strconv"
)

var (
	ErrWriteTooLong = errors.New("write too long")
)

// Writer provides sequential writing of a ar archive.
//
// Write.WriteMagicBytes begins a new file, then WriteHeader begins a
// new file with the provided Header, and then Writer can be treated as
// an io.Writer to supply that file's data.
type Writer struct {
	io io.Writer
	n  int64
}

// NewWriter creates a new Writer writing to w.
func NewWriter(w io.Writer) *Writer {
	return &Writer{io: w}
}

// WriteMagicBytes writes ar magic bytes header.
func (w *Writer) WriteMagicBytes() error {
	_, err := w.io.Write([]byte(MagicString))
	return err
}

// WriteHeader writes header and prepares to accept the file's contents.
//
// The Header.Size determines how many bytes can be written for the next
// file. If the current file is not fully written, then this returns an
// error.
func (w *Writer) WriteHeader(header *Header) error {
	buf := make([]byte, HeaderByteSize)

	writeString(buf[0:16], header.Name)
	writeInt(buf[16:28], header.Date.Unix())
	writeInt(buf[28:34], header.Uid)
	writeInt(buf[34:40], header.Gid)
	writeOctal(buf[40:48], header.Mode)
	writeInt(buf[48:58], header.Size)
	writeString(buf[58:60], "`\n")

	_, err := w.io.Write(buf)
	if err != nil {
		return err
	}

	w.n = header.Size
	return nil
}

// Write writes to the current file in the ar archive.
//
// Write returns the error ErrWriteTooLong if more than Header.Size
// bytes are written after WriteHeader.
func (w *Writer) Write(b []byte) (int, error) {
	if int64(len(b)) > w.n {
		return 0, ErrWriteTooLong
	}

	if len(b)%2 == 1 {
		b = append(b, '\n')
	}

	n, err := w.io.Write(b)
	return n, err
}

func writeString(b []byte, s string) {
	for len(s) < len(b) {
		s = s + " "
	}

	copy(b, []byte(s))
}

func writeInt(b []byte, i int64) {
	s := strconv.FormatInt(i, 10)
	for len(s) < len(b) {
		s = s + " "
	}

	copy(b, []byte(s))
}

func writeOctal(b []byte, i int64) {
	s := strconv.FormatInt(i, 8)
	for len(s) < len(b) {
		s = s + " "
	}

	copy(b, []byte(s))
}
