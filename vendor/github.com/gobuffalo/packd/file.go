package packd

import (
	"bytes"
	"io"
	"os"
	"time"
)

var _ File = virtualFile{}

type virtualFile struct {
	*bytes.Buffer
	Name string
	info fileInfo
}

func (f virtualFile) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (f virtualFile) FileInfo() (os.FileInfo, error) {
	return f.info, nil
}

func (f virtualFile) Close() error {
	return nil
}

func (f virtualFile) Readdir(count int) ([]os.FileInfo, error) {
	return []os.FileInfo{f.info}, nil
}

func (f virtualFile) Stat() (os.FileInfo, error) {
	return f.info, nil
}

// NewDir returns a new "virtual" file
func NewFile(name string, r io.Reader) (File, error) {
	bb := &bytes.Buffer{}
	io.Copy(bb, r)
	return virtualFile{
		Buffer: bb,
		Name:   name,
		info: fileInfo{
			Path:     name,
			Contents: bb.Bytes(),
			size:     int64(bb.Len()),
			modTime:  time.Now(),
		},
	}, nil
}

// NewDir returns a new "virtual" directory
func NewDir(name string) (File, error) {
	bb := &bytes.Buffer{}
	return virtualFile{
		Buffer: bb,
		Name:   name,
		info: fileInfo{
			Path:     name,
			Contents: bb.Bytes(),
			size:     int64(bb.Len()),
			modTime:  time.Now(),
			isDir:    true,
		},
	}, nil
}
