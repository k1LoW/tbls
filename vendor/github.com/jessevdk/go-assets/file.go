package assets

import (
	"bytes"
	"os"
	"path"
	"time"
)

// An asset file.
type File struct {
	// The full asset file path
	Path string

	// The asset file mode
	FileMode os.FileMode

	// The asset modification time
	Mtime time.Time

	// The asset data. Note that this data might be in gzip compressed form.
	Data []byte

	fs       *FileSystem
	buf      *bytes.Reader
	dirIndex int
}

// Implementation of os.FileInfo

func (f *File) Name() string {
	return path.Base(f.Path)
}

func (f *File) Mode() os.FileMode {
	return f.FileMode
}

func (f *File) ModTime() time.Time {
	return f.Mtime
}

func (f *File) IsDir() bool {
	return f.FileMode.IsDir()
}

func (f *File) Size() int64 {
	return int64(len(f.Data))
}

func (f *File) Sys() interface{} {
	return nil
}

// Implementation of http.File

func (f *File) Close() error {
	f.buf = nil
	f.dirIndex = 0

	return nil
}

func (f *File) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *File) Readdir(count int) ([]os.FileInfo, error) {
	if f.IsDir() {
		ret, err := f.fs.readDir(f.Path, f.dirIndex, count)
		f.dirIndex += len(ret)

		return ret, err
	} else {
		return nil, os.ErrInvalid
	}
}

func (f *File) Read(data []byte) (int, error) {
	if f.buf == nil {
		f.buf = bytes.NewReader(f.Data)
	}

	return f.buf.Read(data)
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	if f.buf == nil {
		f.buf = bytes.NewReader(f.Data)
	}

	return f.buf.Seek(offset, whence)
}
