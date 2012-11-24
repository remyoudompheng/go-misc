// Package zipfs implements the http.FileSystem interface
// for zip archives.
package zipfs

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func NewZipFS(z *zip.Reader) http.FileSystem {
	return &zipFS{zip: z, cache: make(map[string][]byte)}
}

type zipFS struct {
	zip   *zip.Reader
	cache map[string][]byte
}

var _ http.FileSystem = new(zipFS)

func (fs *zipFS) Open(name string) (http.File, error) {
	// name will be "/" or "/file/name"
	for _, entry := range fs.zip.File {
		if entry.Name == name[1:] {
			f, err := entry.Open()
			if err != nil {
				return nil, err
			}
			data, err := ioutil.ReadAll(f)
			if err != nil {
				return nil, err
			}
			z := &zipFile{Info: entry.FileHeader, Data: bytes.NewReader(data)}
			return z, nil
		}
	}
	return nil, os.ErrNotExist
}

type zipFile struct {
	Info zip.FileHeader
	Data *bytes.Reader
}

func (f *zipFile) Close() error                              { return nil }
func (f *zipFile) Stat() (os.FileInfo, error)                { return f.Info.FileInfo(), nil }
func (f *zipFile) Readdir(count int) ([]os.FileInfo, error)  { return nil, os.ErrInvalid }
func (f *zipFile) Read(s []byte) (int, error)                { return f.Data.Read(s) }
func (f *zipFile) Seek(off int64, whence int) (int64, error) { return f.Data.Seek(off, whence) }

var _ http.File = new(zipFile)

type zipDir struct {
	Dirname string
	Files   []*zip.File
}

func (f *zipDir) Close() error                              { return nil }
func (f *zipDir) Stat() (os.FileInfo, error)                { return (*zipDirInfo)(f), nil }
func (f *zipDir) Readdir(count int) ([]os.FileInfo, error)  { return nil, nil }
func (f *zipDir) Read(s []byte) (int, error)                { return 0, os.ErrInvalid }
func (f *zipDir) Seek(off int64, whence int) (int64, error) { return 0, os.ErrInvalid }

type zipDirInfo zipDir

func (d *zipDirInfo) Name() string       { return filepath.Base(d.Dirname) }
func (d *zipDirInfo) Size() int64        { return 0 }
func (d *zipDirInfo) Mode() os.FileMode  { return 0600 }
func (d *zipDirInfo) ModTime() time.Time { return time.Unix(0, 0) }
func (d *zipDirInfo) IsDir() bool        { return true }
func (d *zipDirInfo) Sys() interface{}   { return (*zipDir)(d) }

var _ os.FileInfo = new(zipDirInfo)
