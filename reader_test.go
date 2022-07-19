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
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReader(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	f, err := os.Open("testdata/apt_2.4.5_amd64.deb")
	require.NoError(err)

	_, err = NewReader(f)
	assert.NoError(err)

	f, err = os.Open("testdata/badfile.txt")
	require.NoError(err)

	_, err = NewReader(f)
	assert.Error(err)
}

func TestReadHeader(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	f, err := os.Open("testdata/apt_2.4.5_amd64.deb")
	require.NoError(err)

	r, err := NewReader(f)
	assert.NoError(err)

	hdr, err := r.Next()
	assert.NoError(err)
	assert.Equal("debian-binary", hdr.Name)
	assert.Equal(int64(0), hdr.Uid)
	assert.Equal(int64(0), hdr.Gid)
	assert.Equal(int64(33188), hdr.Mode)
	assert.Equal(int64(4), hdr.Size)

	hdr, err = r.Next()
	assert.NoError(err)
	assert.Equal("control.tar.xz", hdr.Name)
	assert.Equal(int64(0), hdr.Uid)
	assert.Equal(int64(0), hdr.Gid)
	assert.Equal(int64(33188), hdr.Mode)
	assert.Equal(int64(6584), hdr.Size)

	hdr, err = r.Next()
	assert.NoError(err)
	assert.Equal("data.tar.xz", hdr.Name)
	assert.Equal(int64(0), hdr.Uid)
	assert.Equal(int64(0), hdr.Gid)
	assert.Equal(int64(33188), hdr.Mode)
	assert.Equal(int64(1489240), hdr.Size)

	_, err = r.Next()
	assert.ErrorIs(err, io.EOF)
}

func TestRead(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	f, err := os.Open("testdata/apt_2.4.5_amd64.deb")
	require.NoError(err)

	r, err := NewReader(f)
	assert.NoError(err)

	hdr, err := r.Next()
	assert.NoError(err)
	data := make([]byte, hdr.Size)
	i, err := r.Read(data)
	assert.NoError(err)
	assert.Equal(hdr.Size, int64(i))

	hdr, err = r.Next()
	assert.NoError(err)
	data = make([]byte, hdr.Size)
	i, err = r.Read(data)
	assert.NoError(err)
	assert.Equal(hdr.Size, int64(i))

	hdr, err = r.Next()
	assert.NoError(err)
	data = make([]byte, hdr.Size)
	i, err = r.Read(data)
	assert.NoError(err)
	assert.Equal(hdr.Size, int64(i))

	_, err = r.Next()
	assert.ErrorIs(err, io.EOF)
}
