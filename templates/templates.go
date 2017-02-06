// Code generated by go-bindata.
// sources:
// templates/base/footer.tmpl
// templates/base/head.tmpl
// templates/chaos.tmpl
// templates/index.tmpl
// templates/public/style.css
// templates/templates.go
// templates/test.tmpl
// DO NOT EDIT!

package templates

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _baseFooterTmpl = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\xb2\xd1\x4f\xc9\x2c\xb3\xe3\xb2\xd1\x4f\xca\x4f\xa9\x04\xd1\x19\x25\xb9\x39\x76\x80\x00\x00\x00\xff\xff\x27\xad\x80\x0f\x16\x00\x00\x00")

func baseFooterTmplBytes() ([]byte, error) {
	return bindataRead(
		_baseFooterTmpl,
		"base/footer.tmpl",
	)
}

func baseFooterTmpl() (*asset, error) {
	bytes, err := baseFooterTmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "base/footer.tmpl", size: 22, mode: os.FileMode(420), modTime: time.Unix(1485783225, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _baseHeadTmpl = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x9c\xcf\x31\x6a\xc5\x30\x0c\x06\xe0\xdd\xa7\x10\xde\x6b\xc1\x1b\x3a\x14\xc7\x77\xf1\x93\x55\xac\xd4\x71\x82\xa5\x04\x72\xfb\xd2\xa4\xb4\xd0\xb1\xa3\x24\xf4\xfd\xfc\xb1\xda\xd2\x92\x8b\x95\x73\x49\x0e\x20\x36\xe9\x1f\x50\x07\xbf\x4f\x1e\xb7\xfd\xd9\x84\x50\xed\x6c\x1c\x48\xd5\xc3\xe0\x36\xf9\x6b\xd6\xca\x6c\x1e\x7f\x7f\xfe\x9e\xbe\x91\x6a\xb6\xe9\x1b\x22\x95\x1e\x66\x2d\xdc\xe4\x18\xa1\xb3\xa1\xf2\x92\xbb\x09\xbd\xec\x82\x8f\xf0\x08\xaf\x3f\x9b\xb0\x48\xbf\xe2\x2e\x5c\x69\xc8\x66\xa0\x83\xfe\x8f\xcd\xea\x53\xc4\x5b\x4a\x2e\xe2\xdd\x36\x3e\xd7\x72\x7e\x65\x00\xc4\x22\x07\x50\xcb\xaa\x93\xdf\x05\x68\xed\x96\xa5\xf3\xf0\xc9\x7d\x06\x00\x00\xff\xff\xbd\xc2\x56\x5d\x22\x01\x00\x00")

func baseHeadTmplBytes() ([]byte, error) {
	return bindataRead(
		_baseHeadTmpl,
		"base/head.tmpl",
	)
}

func baseHeadTmpl() (*asset, error) {
	bytes, err := baseHeadTmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "base/head.tmpl", size: 290, mode: os.FileMode(420), modTime: time.Unix(1485783225, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _chaosTmpl = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\xac\x93\x41\x6b\xe3\x30\x10\x85\xef\xfe\x15\x83\xc9\xd9\x22\x39\x2e\x8a\x60\x59\xf6\xba\x2c\x69\xe9\x5d\xb1\xc6\x91\xa8\x2c\x1b\x69\x1c\x28\x42\xff\xbd\x48\x76\x88\xdb\xb4\x21\x87\xfa\x64\xe6\x8d\xdf\x9b\xf9\x2c\xc5\x48\xd8\x8f\x56\x12\x42\x7d\x94\x01\x99\x46\xa9\x6a\x68\x52\xaa\x2a\xae\xb7\xe2\x8f\x96\x43\x80\x7e\x70\xaf\xf8\xc6\x99\xde\x8a\xaa\xe2\xa3\xa8\x00\x00\xd6\x12\x98\x00\x31\x36\xa5\xd4\x3c\x91\x24\x4c\xa9\x29\x5d\x31\x9a\x0e\x16\xe1\x77\x4b\xe6\x8c\x29\x15\x21\x3f\x5c\x82\xf6\xd8\xed\x6b\xd6\xe6\x06\x36\xca\x29\x60\x0d\xad\x95\x21\xec\xeb\xc9\x40\x6f\x9c\x01\x6f\x4e\x9a\xa0\xb3\x83\x24\x54\x70\x9c\x88\x06\x57\x8b\xff\xb9\x97\x33\x29\x96\x18\xb4\xe1\x9e\xb5\xc7\x30\xf5\x8f\x7a\x1f\x4a\xf3\xda\xdc\xa9\x94\x2a\xce\xc6\xbc\xbf\xde\x89\xbc\xa2\x09\x64\xda\xc0\x99\xde\xe5\x22\xc9\xa3\xc5\x95\x7b\x8b\xd6\xa2\x82\x40\xde\x8c\xa8\xa0\xc8\xf5\x6c\xc7\x29\x43\xbe\xbc\x7b\x71\x9d\x99\xf4\xd5\x99\x33\xd2\x1f\xa5\x17\x69\x27\xbc\x96\x39\xbb\x7c\x9b\x6b\xc5\x31\x46\xf0\xd2\x9d\x10\x36\x81\xe0\xd7\x1e\x56\x3f\x64\x9e\x16\x16\x42\x9f\x62\x95\x88\x71\x13\xa8\xf9\x27\x7b\x4c\x89\x33\x52\x5f\xa9\x25\x7f\x2d\xcf\x13\xc4\x08\xe8\x14\x14\x3e\x65\xcd\x85\xd1\x01\x5b\x74\x04\x05\xff\x4f\x53\x7a\x36\x3d\xde\x02\xca\xe7\x6b\x70\x0f\x12\xc2\x15\xa0\xbf\x67\x74\x74\x17\x0e\x36\x39\xf2\x1b\x36\xd8\xcc\xc9\x8f\xb1\xb9\xb9\x72\xdd\x30\x10\xfa\x72\xe9\xde\x03\x00\x00\xff\xff\x80\xb0\x0c\xf4\x8f\x03\x00\x00")

func chaosTmplBytes() ([]byte, error) {
	return bindataRead(
		_chaosTmpl,
		"chaos.tmpl",
	)
}

func chaosTmpl() (*asset, error) {
	bytes, err := chaosTmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "chaos.tmpl", size: 911, mode: os.FileMode(420), modTime: time.Unix(1486047848, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _indexTmpl = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\xb4\x56\x5f\x6f\xa4\x36\x10\x7f\xdf\x4f\x31\x42\xfb\x5a\xac\xbb\xc7\x13\x41\x4a\xb3\x3d\x5d\xd4\xdc\xf5\xb4\x49\xd4\x67\x2f\x0c\x8b\x55\x83\x91\x3d\xe4\x5a\x59\x7c\xf7\xca\xc6\x70\xb0\x90\x88\xa8\xdb\x7d\x62\xe7\x9f\x67\xe6\x37\xf3\xb3\xad\x25\xac\x1a\xc9\x09\x21\x3a\x71\x83\xac\x44\x9e\x47\x10\x77\xdd\x6e\x97\x94\x1f\xd2\x3f\x51\x66\xaa\x42\x20\x05\x4f\x68\xe8\xf6\x8c\x35\x25\xac\xfc\x90\xee\x92\x26\xdd\xdd\x1f\x3e\x81\xb5\xf1\xfd\xa1\xeb\x76\xc9\x49\xb3\x74\xf7\xdc\x78\xc9\x73\x43\xa2\x42\x27\x65\x4d\xea\x22\x7d\x4c\xef\x64\x6b\x08\x75\xc2\xca\x8f\x4e\x42\xfc\x24\x11\x32\xc9\x8d\xb9\x89\x5a\x01\x19\x4a\x89\x39\x18\xd2\xa2\xc1\x1c\xbc\x3a\x4a\x77\x00\x00\x09\xb9\xa4\x86\x6f\xdd\x7f\x04\x45\x7a\x7f\x48\x18\x95\x73\xd9\x9d\x46\x4e\x98\xb3\x47\xe2\x9a\x30\x4f\x18\xe5\x73\x83\x50\xc6\xc2\x4f\x29\x9d\x8b\x9a\x93\xd2\x4b\xe5\xe1\xd7\x47\xd4\x2f\x38\xd1\x24\x6c\x48\xc6\xc9\x7c\x8a\xd6\x82\xe6\xf5\x19\x61\x5f\xc1\xa7\x1b\x88\xbf\xf2\xac\x14\x35\x1a\xe8\xba\xb5\xf4\x27\x69\xb9\x9f\xb5\xfb\xaa\x6f\xe6\x54\x9a\x70\x28\x35\x16\x37\x11\x93\xea\x6c\xd8\x68\xc4\xaa\x3e\x78\x04\x24\x48\xe2\x4d\xf4\xa0\xce\x26\x4a\x13\x31\x74\xb5\x10\x12\x81\xf0\x6f\x02\xd5\x92\x14\x35\x82\xc8\x54\x1d\xa5\x09\x13\x69\xc2\x78\xba\xed\x94\x1a\xe9\x87\xd2\x7f\x8d\xa7\x7c\xeb\xff\x83\xbc\x38\xed\x87\x28\xc4\x6b\x07\x5c\x22\xb0\x56\x77\x00\xed\x96\x2e\xca\x67\x4b\xd3\x80\xeb\xcc\x74\x7e\x84\xb5\xa2\x80\x7d\x15\x7f\xe1\xc6\x63\x3d\x35\xa4\x7c\x48\xd9\x5a\xc8\x8c\x39\x22\xcf\xff\x71\xc6\x0f\x3c\x0c\xb8\x97\x3c\x12\xa7\xd6\x01\x17\xad\x37\xca\x5a\xe7\xe3\xed\x9f\x8f\x0f\xd0\x75\xc3\x54\x6d\x6d\x2c\x77\xe6\xff\x1d\xbc\xcb\xc2\x51\x1a\x9c\x97\x9b\xfe\xb2\xb0\xa9\xf3\xcd\x1d\x99\x2c\xc5\x3b\xfa\x32\xf1\x0a\xdd\x99\x2d\xd7\xd6\x1e\x65\x3f\x9d\xae\xdd\xa9\xb7\xab\x1e\xb6\xfd\x1d\x25\x0f\x2e\xa1\xde\x9f\x7c\xb1\xb5\xd8\xfc\x64\xbc\xc7\x35\x2b\xed\x39\xca\x5a\xc0\x3a\x07\xcf\xc7\x9e\x59\x03\x27\x3b\x4a\x37\xd7\x66\xe4\x6f\xbc\xc2\x25\x7d\xba\x1e\xae\x88\x6f\x33\x12\xaa\x36\x1b\x49\x95\x3c\xa9\xfa\xac\x27\x8c\x3a\x87\xd1\x69\xff\xf8\x1d\xf6\x14\x7f\xe6\x42\xb6\x1a\x2f\x80\x73\x0b\x31\x42\x40\x68\xc8\x21\x40\xb1\xcb\xda\xd9\x4d\xfe\x2c\x59\xcb\x2b\x87\xb0\x5d\x07\x45\xf8\x7c\xc5\x32\x14\x37\x8d\xf4\x36\x1e\x9f\xc7\x78\xd7\x85\xe4\x49\xac\x41\xf2\x15\x8d\xe1\xe7\x15\xc5\x11\x1b\xa5\x69\x23\x26\xda\x63\xd2\xbb\xbc\x7e\xcf\x59\xbb\xd7\xf1\x93\x78\xb5\xad\x3a\x0e\xd9\xac\xe9\x47\xb8\xbc\xe1\x97\x23\x16\x01\x29\x1d\x7f\xe7\x54\x3a\x17\x9e\x6e\xed\xf1\x11\x33\xac\x09\xb2\x92\xab\xd0\x67\xf7\x92\x71\x7e\xfd\x9a\xfb\xf7\xcb\x9d\xd3\xc6\x7e\x66\x43\x45\xfe\x46\x09\x72\x87\xeb\xcb\x8c\x64\xc7\x95\xf6\x61\x59\xc3\x5b\x83\xd1\x04\xb8\x4a\xd4\x02\xb4\x38\x97\x04\x85\x54\xee\x9e\x83\x53\x4b\xe4\x96\xf7\xbb\xb3\x1d\xb7\x77\xc9\xdf\x17\xa1\x35\x9a\xb6\xda\x1a\xfb\xe8\x8d\xa7\xc1\x47\xe2\xef\x9f\x6b\x2b\x27\x44\xe9\x01\x89\x0b\x69\xbc\x5b\x78\xbf\xfd\xbf\x83\xd8\xef\xc9\xc6\x79\x43\x3f\x6f\x3d\x10\xbf\xbd\x60\xfd\xf6\xd0\xe1\x5b\x43\x87\x61\x43\xb7\x2d\xe8\xe2\xa9\x5c\x28\x45\x8e\xad\xe3\xae\xfb\x37\x00\x00\xff\xff\x58\x5a\xc6\xee\x47\x0b\x00\x00")

func indexTmplBytes() ([]byte, error) {
	return bindataRead(
		_indexTmpl,
		"index.tmpl",
	)
}

func indexTmpl() (*asset, error) {
	bytes, err := indexTmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "index.tmpl", size: 2887, mode: os.FileMode(420), modTime: time.Unix(1486396928, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _publicStyleCss = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x74\xcc\x31\x0a\xc3\x30\x0c\x85\xe1\x3d\xa7\x78\x63\x0b\xae\x31\x81\x2c\xed\x69\x94\x44\x15\x25\x41\x02\xc5\x19\x4a\xe8\xdd\x8b\xa9\x0d\x5d\x32\x49\xc3\xf7\xbf\xe8\x4c\xf3\x3b\xe0\x77\x41\x38\x30\xd9\x6a\x7e\x87\x38\xb3\x3e\xf0\xe9\xa2\x5a\x6e\xaa\xbd\xff\xd0\x9c\x54\xb8\xc8\x2e\x66\xde\xb2\x2d\x38\x30\xd2\xb4\x88\xdb\xae\xf3\xad\x32\x97\x91\x2e\x29\xf4\xc3\x10\x52\x48\xb1\xbf\x96\xe9\xe2\x9f\xf4\x5a\x77\xe7\xed\xbc\xaa\x4d\xab\xbe\x01\x00\x00\xff\xff\x74\xca\x77\x00\xb4\x00\x00\x00")

func publicStyleCssBytes() ([]byte, error) {
	return bindataRead(
		_publicStyleCss,
		"public/style.css",
	)
}

func publicStyleCss() (*asset, error) {
	bytes, err := publicStyleCssBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "public/style.css", size: 180, mode: os.FileMode(420), modTime: time.Unix(1486368821, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _templatesGo = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\xc4\x99\xd9\x6f\x23\xc7\x11\xc6\x9f\x39\x7f\xc5\x58\x80\x0d\x32\x50\xa4\xb9\x0f\x01\xfb\xe2\x0b\xf1\x43\x6c\x20\xd9\x3c\xa5\x03\xa3\x67\xa6\x47\x26\x22\x89\x0a\x49\x39\xbd\xbb\xd8\xff\x3d\xf8\x55\x15\x29\x4a\x2b\x5f\x8b\x05\xf2\xc0\x15\x39\x47\x77\x1d\x5f\x7d\x5f\x55\xef\xe5\x65\xfa\xd5\x66\x0a\xe9\x75\xb8\x0b\x5b\xbf\x0f\x53\x3a\xbc\x49\xaf\x37\x7f\x1e\xd6\x77\x93\xdf\xfb\x8b\xe4\xf2\x32\xdd\x6d\x1e\xb6\x63\xd8\x5d\xf1\x7d\x1f\x6e\xef\x6f\xfc\x3e\xec\x2e\x07\xbf\x0b\x97\xf3\x66\xb3\x0f\xdb\x8b\xfd\xed\xfd\xcd\x0b\x77\x7f\x0a\x7e\x7a\xe1\xde\xf8\x93\xdf\xec\x5e\xb8\xbe\xbe\x9b\x42\x7c\xe1\xfa\xfd\xc3\x70\xb3\x1e\x2f\x77\xfb\x37\x37\xe1\x62\xdc\xed\x9e\xde\x3d\x7e\xbb\xb8\xde\x3c\xbf\xb3\xdb\x1f\x97\xfb\xfa\x87\xf4\xfb\x1f\x5e\xa7\xdf\x7c\xfd\xdd\xeb\xcf\x92\xe4\xde\x8f\xff\xf6\xd7\xe1\xf1\xe1\x24\x59\xdf\xde\x6f\xb6\xfb\x74\x99\x2c\xce\x86\x37\xfb\xb0\x3b\x4b\x16\x67\xe3\xe6\xf6\x7e\x1b\x76\xbb\xcb\xeb\xb7\xeb\x7b\x2e\xcc\xb7\x7b\xfe\xac\x37\xfa\xef\xe5\x7a\xf3\xb0\x5f\xdf\xf0\x63\x23\x2f\xdc\xfb\xfd\x4f\x97\xf3\xfa\x26\xf0\x85\x0b\xbb\xfd\x76\x7d\x77\x2d\xf7\xf6\xeb\xdb\x70\x96\xac\x92\x64\x7e\xb8\x1b\x53\x0b\xf0\xdf\x82\x9f\x96\x7c\x49\xff\xf9\x2f\xb6\x3d\x4f\xef\xfc\x6d\x48\xf5\xb5\x55\xba\x3c\x5c\x0d\xdb\xed\x66\xbb\x4a\xdf\x25\x8b\xeb\xb7\xf2\x2b\xbd\x7a\x95\x62\xd5\xc5\xf7\xe1\xbf\x2c\x12\xb6\x4b\x31\x9b\xdf\x5f\x3e\xcc\x73\xd8\xca\xb2\xab\x55\xb2\x58\xcf\xf2\xc2\x67\xaf\xd2\xbb\xf5\x0d\x4b\x2c\xb6\x61\xff\xb0\xbd\xe3\xe7\x79\x3a\xdf\xee\x2f\xbe\x61\xf5\x79\x79\xc6\x42\xe9\xe7\xff\xb9\x4a\x3f\xff\xf9\x4c\x2d\x91\xbd\x56\xc9\xe2\x7d\x92\x2c\x7e\xf6\xdb\x74\x78\x98\x53\xdd\x47\x37\x49\x16\x3f\xaa\x39\xaf\xd2\xf5\xe6\xe2\xab\xcd\xfd\x9b\xe5\x17\xc3\xc3\x7c\x9e\x5e\xbf\x5d\x25\x8b\xf1\xe6\x9b\x83\xa5\x17\x5f\xdd\x6c\x76\x61\xb9\x4a\x3e\x95\x3d\x2c\xa3\xeb\xff\xc2\x42\x61\xbb\x55\xbb\xed\xe2\xf0\x30\x5f\x7c\x89\xe9\xcb\xd5\x39\x4f\x24\xef\x93\x64\xff\xe6\x3e\xa4\x7e\xb7\x0b\x7b\x42\xfe\x30\xee\x59\x45\xfc\xb3\x7c\x24\x8b\xf5\xdd\xbc\x49\xd3\xcd\xee\xe2\xdb\xf5\x4d\xf8\xee\x6e\xde\x1c\xdf\xb3\x14\x1e\xae\x9f\xac\x20\x39\x4c\x53\x4b\x63\xb2\xd8\xad\xdf\xca\xef\xf5\xdd\xbe\xa9\x92\xc5\x2d\x15\x97\x1e\x17\xfd\xeb\x66\x0a\x72\xf1\xf5\xfa\x36\xa4\xc0\xe4\x82\x6f\xec\x23\x50\x59\xce\xeb\xe7\x7b\xad\xd2\xef\xfd\x6d\x58\xae\x6c\x07\xf6\x34\x2f\xe7\xf5\x05\xbb\x27\xef\x7f\xe5\xdd\xbf\xaf\xdf\xf2\xae\x58\xf3\xf4\x55\x0c\xfd\xd5\x57\xb1\x75\xb9\x3a\xb5\xfc\xe9\x02\xb8\xf6\x5b\x0b\xe0\xdc\x72\xf5\xe8\xe8\x07\x2b\x98\xf7\xbf\xbc\xc8\x77\xbb\xaf\xd7\xdb\xe5\x2a\x1d\x36\x9b\x9b\xd3\xb7\xfd\xcd\xee\x37\x3c\x7f\xb3\x53\xc7\xc3\x76\xf6\x63\x78\xf7\xfe\xe4\x6d\x83\x04\x28\xff\x11\xfa\xfa\x56\xb8\xed\xf5\xed\xfd\x4d\xfa\xca\xd0\xb0\x3c\x73\x31\x9f\x5d\xec\x06\x17\xb3\xce\xc5\x2c\xb3\x4f\xef\x62\x13\x5c\xec\xec\xda\x3c\xbb\x38\x14\x2e\x4e\xb9\x8b\xd5\xec\xe2\xd8\xbb\x58\x8c\x2e\x0e\xa5\x8b\xa1\x7c\x76\xcf\xeb\x5f\xdf\xbb\x98\x55\x7a\x3d\xe7\xf9\xda\xc5\xa1\x77\xb1\xec\x5d\x6c\x1b\x17\xbb\xec\x64\x3f\xdb\x83\x4f\xd1\xba\xe8\x27\xbb\x3f\xbb\x98\x37\xa7\xcf\x9d\x1d\xf9\xe6\x89\x47\x56\x07\x2f\xf1\xcb\xa1\x5a\x4e\xf8\x29\x59\x2c\x9e\x45\xe4\x3c\x59\x2c\xce\x9e\x2b\xc0\xd9\x79\xb2\x58\x1d\x51\xfb\xf4\x05\xf6\xfa\x93\x14\xda\xe9\x5e\x52\x69\x47\x3a\x7b\xd1\xc4\xdf\xa2\x8b\x63\x95\x4b\x9d\xb2\xca\xd3\x9c\xbf\xa3\x1a\xae\xd2\x17\x8c\x4d\x01\xfb\x55\x5a\x14\xe7\x29\xa8\xbd\x3a\x05\xf5\xb2\x2a\xb2\x95\x5c\x07\x8b\x57\x8a\xd5\x7f\xdc\xad\xe3\x32\xaf\xba\xba\xed\xca\xa2\xa8\xcf\xd3\x6c\xf5\x3e\x59\x78\x36\xfd\x42\x5c\x7b\x27\xfe\x5c\xa5\xe6\x16\x16\x5d\xc9\xbf\xef\x8f\x61\xf5\xe7\x1f\xe0\xec\x2f\xc1\x4f\x1f\x8d\xb2\x7e\x74\x71\x9c\x5d\x2c\x73\x17\x1b\xef\xe2\x58\xbb\x58\xf2\xec\xe8\x62\xd6\xb8\x18\x32\x17\xa7\xc9\x45\xdf\xba\x98\xf3\x3d\xb8\xd8\x0c\x2e\x8e\xa0\x6c\x70\xb1\xf4\x2e\xe6\x95\x8b\x63\xeb\x62\xdb\xba\x38\xe7\x2e\xf6\xa5\x8b\x75\xed\xa2\x1f\x5d\x9c\x2a\x17\xdb\xdc\xc5\xae\x70\xd1\xd7\x8a\xd0\xb6\x70\x71\x1e\x5c\x9c\xb8\x56\xb9\x38\x80\xda\xcc\xc5\x21\x77\xd1\x97\x2e\x16\x95\x8b\x33\x9f\xc9\xc5\x79\xd4\xeb\x93\xd7\xe7\xfb\x42\xfd\xea\x6b\x17\xdb\xd2\xc5\x0a\xd4\x07\x17\x8b\xcc\xc5\x12\x7b\x7b\xf5\xbd\xc6\x87\xd6\xc5\x61\xd6\xea\xc8\x83\x8b\x43\xab\x6b\x4e\xbd\x8b\x5d\xa5\xcf\x84\xc9\xc5\x66\x74\x31\x1f\x5d\xac\x3a\x17\xa7\xda\xc5\xb1\x54\xbf\x59\x6f\xee\xd5\xdf\xa9\xd1\x4a\x93\x67\x83\x8b\xed\xac\x9f\x39\xb8\xd8\xb3\x36\x7f\x2d\x86\xc3\xc1\x8e\xc1\xc5\xa2\x50\x5b\x79\xa7\x69\x5c\x2c\xd8\x7f\x74\x31\x54\x2e\xe6\x9d\x8b\x3e\xd7\xaa\xe6\xef\x6c\xfe\x0d\x83\xe6\x6a\x98\x5c\x0c\xa3\xc6\xae\x63\xaf\x4c\xf3\xe9\xc9\x17\xb9\x1b\x5c\x1c\x32\xb5\x1b\x3f\x43\xe1\x62\x11\x5c\xac\x47\x17\x9b\xde\xc5\xb1\xd3\x3d\x7d\xe6\x62\x57\xaa\xad\xac\x33\xb2\xae\x77\xb1\x2e\x5d\x1c\x89\xc3\xe0\x62\xe5\xf5\x5d\xd6\x20\xdf\xf8\x5e\x06\x17\xa7\x56\xf3\xd5\x62\x7f\xad\xb8\xe1\x1d\xfc\x22\xbe\xc4\x70\x1c\x5c\xf4\x5e\xf3\x3e\xcd\x2e\x66\xb5\x8b\x4d\xa7\xb1\xed\x1b\xcd\xfb\x5c\xaa\xfd\x30\x59\x3b\x29\xb6\x9e\x33\x11\xfe\x8e\x85\x8b\x75\xe3\x62\x3d\xd9\x1e\xf9\x8b\x4c\x74\xc0\xfc\xc7\xf1\xd0\xe1\xed\x47\x16\x3a\x76\x9a\x1f\x72\xd0\xe1\xe1\xdf\xcb\x40\xcf\x4c\xfb\xa4\xfc\x73\x62\xe6\x81\x7d\xfa\xec\xff\x48\x3f\xd2\x89\x7f\x34\xf7\xc0\x0f\x60\xa6\xca\xb5\xc6\x50\x36\xb8\x07\x9e\xe9\x6a\x17\x83\xd5\x57\x5e\x2b\x7e\xc1\x0e\xb5\x0b\x2e\x50\x35\xf0\xda\x51\x93\x99\x8b\x75\xef\xe2\xdc\xb8\x38\x78\x55\x4a\xf0\x4f\x0d\x82\x23\xf8\x63\x6c\xb4\x3e\x7d\xa7\xf7\xa9\x4d\x9e\xa1\xee\x8b\xce\xc5\xaa\x78\xc4\x20\xf5\x24\x8a\x09\x1f\x0c\xca\x4f\x05\x1c\x06\x7f\xb0\x5f\xe5\x62\xe0\xfe\xa4\x78\xa7\x0e\x67\x53\x68\xf8\x53\x78\xa4\xd3\x5a\x03\xcb\x79\xa1\xeb\x83\xfb\xbe\x52\x4c\xf7\xbd\x8b\x55\xa3\xaa\x0d\x67\x50\x2f\x75\xa1\x75\x54\x78\x17\xbd\xf1\x55\xb0\xda\xa7\x8e\xaa\x52\x55\x9a\x5a\x6c\x33\xe5\x81\xb9\x33\xdf\x7a\xe5\x66\xe2\xc1\x1a\xbc\x07\x97\xfa\x93\x1c\x4c\x66\x4b\x6f\xd7\xe0\xfc\xd2\xea\x13\x9e\x2e\x47\x8d\x11\xfc\x5b\x8d\x6a\x5b\x61\x31\xe4\xd9\xde\x2b\xd7\xd2\x21\x84\x5c\xf9\xbe\x1a\x2c\x16\xa3\x3e\x4b\xae\x88\x15\xfc\x54\x1b\x77\x91\x97\xc9\xf8\xe0\xc0\xa5\xbc\x03\x07\x94\xc6\xab\xbc\x4b\x1e\xb3\x49\xbb\x91\xde\xd6\x81\xff\xc2\xa0\xb9\xaf\x32\x17\x1b\xd3\x2c\x62\xc8\xf7\x2a\xa8\x5d\x70\x1c\x3c\xda\x99\x7e\xd4\x95\xc6\x88\x67\x05\x73\x70\x4a\xab\xb8\x94\x3c\xf7\x2e\xf6\xad\x8b\x65\xa9\x7b\x13\x5f\xf8\x98\x5c\xe3\x9b\x70\x7a\xad\xba\x06\x26\xe7\x5a\xf3\xd0\x7a\xc5\x37\x7c\x09\xa7\x89\x2e\x59\xac\xc8\x2d\xb9\x1b\x83\xe6\x64\x18\x95\x97\x25\x2f\xbd\xfe\xf6\x85\xd5\x42\xa5\xef\x8e\xd9\x63\xee\x42\xad\xd8\xc6\x57\x72\x07\x96\x44\xe7\x06\xb5\x9b\xb8\xf1\x3e\x31\xad\x4c\x8b\x89\x3b\x39\xc6\x7f\xf8\x1e\x3b\xd1\x28\xec\xc6\x2f\x30\x83\x36\xb1\x37\xeb\xa0\xc5\xfd\xa4\x7b\xa2\xb1\xe8\x1d\xdf\x89\x03\xfa\x8a\x6f\xe0\x21\x6f\xb5\x3e\xe8\x0e\x79\x97\xf8\x50\x73\x68\x53\x3b\x2a\xe6\xa5\x0f\xa8\x15\x1b\x60\x19\x7e\x17\x8d\x9e\xb4\x86\xc9\x6d\x97\x5b\x0f\xd1\xaa\x06\x52\x0f\x60\x03\xff\x7b\xc3\x12\x1f\xea\xa3\xe7\xba\xe9\x25\x39\xc3\xd6\xdc\x30\x3c\x9b\x2d\xf8\x00\x36\x3b\xb3\x99\x9a\xa9\xad\xfb\xa5\xe3\x45\x93\xd1\x58\x7a\x9a\xdc\xf4\x0d\x1e\x0a\x9d\xf6\x29\x7c\x4a\xd3\x32\x74\x10\x9e\xa1\x63\x06\x4f\xe8\x60\x5b\x69\x8f\x82\x86\xcd\xd6\x3b\xa0\x93\xe4\x5d\x34\x71\xd2\xb8\x67\x85\xe2\x36\xb4\x1a\x87\xf6\xd0\x3d\x17\xaa\x61\xd4\x81\x60\x12\x9f\x5a\x5d\x97\x0f\x71\xa5\x86\x4a\x8b\x27\x9a\x2f\x7d\x94\xd5\xc8\x38\x6a\xfe\xc1\x1a\xbc\x05\x4e\xf9\xb4\x07\x5d\x36\x7e\x84\x87\xb8\x16\x0c\x3b\x59\xf9\xa1\x9e\xc2\x15\x83\xf5\x72\xe4\x9a\x35\x1f\x9f\x3b\xea\xe9\x91\xc4\xff\xb8\x98\x1e\x5f\x15\x25\x7d\x3c\x97\x79\xaa\xa2\xc7\xa7\x7e\x8f\x84\x3e\xb7\xe6\x53\xe9\xe7\xa9\x71\xa6\x9d\x7d\x9e\xff\x51\xed\x6c\xb2\xaa\xed\xaa\xee\x53\x68\xa7\x9c\x56\x7d\xfc\x74\x58\x59\x9f\x34\x2b\x0f\xd2\x47\x83\x21\xb0\x41\x0d\x4c\x86\x5d\xa9\x4d\xeb\xb7\x6b\xaf\x9a\x4b\x0d\xc3\x0f\xb9\x69\x2f\x5c\x06\x77\x82\x6d\xb8\x1e\x2d\x80\xd3\xe0\x12\xf6\xa1\xc7\xe6\x1a\x38\x2e\x66\xc5\x13\x36\xd6\xa6\xcb\x68\x06\xef\xd2\xd3\xb2\x07\x5a\x0c\x47\xcc\xad\xd6\x08\x5c\xd8\x1a\x16\xfb\x4c\xfd\xf0\xa6\xad\x68\x22\x35\x4f\x1d\xf5\x56\x2b\x68\x43\xd9\x2a\xbf\x4a\x6f\x3c\x69\x6d\x63\x7b\xee\xb5\x3e\x88\x89\xe8\x8c\xb7\xf9\xa2\xd4\xfb\x95\xf1\x4a\x65\xf3\x0a\x1a\x45\xdd\x10\x3f\xf8\x1e\x8e\x26\xc6\xf0\x94\x68\x54\x6e\x3d\xb2\xd7\x38\x31\x47\x50\xfb\xc4\xae\x31\xde\x98\x4d\xdf\xca\xfa\xd1\x0e\xe6\x12\x7c\x09\x41\x7b\xf6\xa2\xd1\x7d\xf8\xb0\x3e\x1c\x03\xf7\xa1\x21\xcc\x44\xcc\x1d\xf4\xf8\x68\x19\xbd\x85\xe8\x59\xaf\xfe\xc1\x0b\xc4\x8a\x98\xb7\x9d\xf5\xc8\xa5\x72\xb7\xb7\xde\x01\xee\x83\xb3\x2a\xeb\xdd\x0b\xeb\x27\xe0\xc2\x60\x7c\xda\x18\x6e\xda\x5a\x39\x88\xd8\x77\xa6\x91\x82\x81\x51\xb5\x31\x27\x6e\xa6\x33\xf4\x30\xf0\x0b\x7c\x4a\xee\x32\x3b\x3d\x00\x13\x9d\x9d\x1e\xa0\x8b\x32\xab\x15\x36\xf7\x8d\xaa\x09\xc4\x9e\xe7\x88\xf1\x29\x56\x59\x0f\x3c\x76\x8d\x62\x93\x98\x80\x49\x78\xb5\xaa\x95\x4b\xf9\x4d\x9f\xc4\xda\x87\x59\x8e\xde\x0e\xed\x40\x4b\x44\xd7\x3b\x8d\x4f\x65\x7d\x10\x3a\x0f\xff\xc9\x3c\xd7\xa9\x5e\xb6\xa6\xf9\x32\x87\x34\xea\x37\xf7\x0b\x5b\x4b\x66\xc5\xc3\x6c\x52\x58\xbf\x52\xab\xc6\xa3\x49\xac\x0d\xb6\xc1\x75\xd9\xa9\x16\x10\xa3\xc6\xe6\xc1\xd1\x72\x05\x9e\xe9\xad\xd0\x60\xf4\x17\x8c\x52\x1f\xc4\xaa\xc9\xb5\x0e\x89\x8d\xcc\xc9\x83\x62\x1c\xee\x87\xdf\x05\xb7\x95\xe2\x07\xfc\x82\x29\xfa\x54\xc1\x46\x65\xb9\xb5\xb8\x51\x03\x72\xdd\x66\x2a\xe2\xc7\x5e\xf4\x33\xf8\x41\x1d\xd1\xaf\xca\x5a\x56\x27\xb9\xcd\xdc\xc1\xe6\x53\x74\x04\xec\xa2\x7d\xac\x01\x16\xc1\x15\xf1\x97\xde\xae\xd3\x78\x82\x25\xf6\x97\xde\x21\x28\x86\xc9\x27\xbc\x00\x6f\x80\x4b\x6c\xa1\x8e\xd0\x35\xf0\x3c\x5b\x0f\x86\xfe\x49\xef\x50\x3c\xfa\x46\x5c\xe8\x41\xb0\x35\x9f\x6c\xde\xef\x75\x3d\x99\x47\x0b\xb5\x03\xdf\x83\xf5\xc2\x12\x8f\x5e\x7b\x6b\xf2\x03\x2e\x3a\x3b\x6f\x68\xad\xf7\x13\x2d\x2b\x0d\x73\xd6\x47\xf2\x7c\x61\x3e\x0a\xee\x5b\xed\x37\xa9\x0b\x78\x06\x6e\x3c\xe0\xf2\x30\x87\x32\x2f\xc8\x0c\x6f\x7d\xeb\x64\xfd\xb4\xf4\x97\xa3\x69\x74\xd0\xb3\x01\xfc\x96\x5e\x2b\xd3\x3a\xcb\xad\xb7\x08\x36\x4b\x34\xc6\x97\xf8\x0d\xd6\xdb\xc3\xf9\x81\xd5\x11\x71\x10\xce\xb0\x79\x65\x34\x4d\x6e\xed\x7d\x7a\x05\xfa\x6d\xb8\x74\xe8\xac\x17\x1d\x35\xce\xf8\x0e\xce\xa4\x16\xed\x74\xae\x34\x2e\xad\xad\xe7\x60\x56\x12\xbf\x66\xd3\xf5\x59\xcf\x50\x5a\x3b\xbf\x20\x07\xcd\xa4\x1f\xe6\x92\xd6\x78\x15\x1b\xc8\x0d\xb1\x64\x6e\x97\x79\x3b\xd7\x3c\xd0\x2b\x0b\x47\x0e\x3a\x77\x0d\x16\x4b\x7a\x2b\xb8\x98\x7e\x29\xb7\xde\xf1\x30\xeb\x94\x83\xf6\x22\xf0\x09\x3c\x97\x79\xad\x77\x79\x66\xd2\x1a\x93\xb9\xc3\x7a\x2e\x39\xb7\x98\x74\xbe\x91\x73\xa5\x5e\xe7\x1c\x3f\xd8\x19\x4d\xa3\x5c\x99\x1b\xe7\xe3\x17\x79\x83\x3b\xe9\x73\xf0\xb7\xb5\x33\x21\x72\x04\xd6\xe1\x35\xf8\x6f\xb0\xbe\x57\x78\x7a\xd0\xf8\x10\xff\x43\xdc\xe8\xb5\xe1\x2c\xb0\x42\x3e\x78\x16\xad\x40\xaf\xe8\x8b\xa8\xc3\xd2\xce\x42\x88\x37\xb6\xa3\x0d\xe8\x1a\x98\x97\xf5\x0b\xeb\x33\x27\xc5\x31\x36\x7b\xd3\x66\x9e\xa1\x27\xcd\x6d\x8e\x13\x3d\x1e\x94\xe7\xe1\xb8\xd2\xe6\x84\xd9\x74\x98\xfd\x3a\xe3\x51\xc1\x7f\xae\x75\x32\xdb\x5c\x47\x0c\xd0\x1e\xea\xab\x35\x1e\x80\xc7\x89\x11\xf9\x2a\xac\xc7\x83\x03\xb3\xe1\x51\xeb\x65\x4e\xea\x14\x83\x68\x64\x65\x67\x6b\xe4\x5b\xe6\x95\xde\x6a\x3f\x68\x1e\xf1\x9f\x7b\xc4\x0c\x3b\xa4\x7e\x26\x8d\x13\x7c\x8d\x2d\xed\x21\xdf\x87\x7a\x6f\xb5\x4e\xe7\x43\xdf\x61\xb9\x91\x73\xc0\x4e\x63\x27\x33\xbc\xf5\x0f\xa2\xcd\x76\xc6\x25\xf3\xfc\xa4\xb5\x04\x96\xf9\xcd\xda\xb9\x9d\x29\xc2\x1b\x32\x7b\x94\xa6\xab\x5e\xeb\x6d\xb2\xb3\x31\x99\x65\x82\xe6\xb0\xb6\x99\x64\x3a\x39\x7b\x14\x8e\xa9\xad\x6e\x6c\x66\x9d\xcc\x06\xb8\x06\xce\xa7\xe6\x58\x47\x66\xbd\x56\xf7\xca\xad\xe6\x64\x8e\x35\xcd\xc2\x3e\x6a\x49\xce\x39\x0d\x2f\xf4\x12\x60\x8b\xfd\x44\x4b\xad\xf7\x10\xed\xb3\x33\x3c\xf0\x4c\x3d\xc2\x35\x68\x3b\xd7\xa8\x7d\xec\x06\xdf\x70\x39\x1c\x2f\x3d\xd2\xa0\x35\x01\x1f\x49\xaf\x64\x67\x85\xe8\xbf\xc4\x3f\xa8\x26\x66\x36\x07\x90\x23\xe1\xac\x41\xf1\xd3\xd8\xf9\x2b\x18\x03\x2f\xa2\x47\xbd\xf1\xc0\xa0\xdc\x29\x33\x4e\xa5\x7e\x83\x6f\x34\x0f\x0e\xa7\xbf\x2b\x8d\x7f\xb8\x4f\xce\xf0\x37\xd8\xb9\xa6\x68\x6b\x69\xba\x59\x29\x8f\x62\xb3\xe4\xa1\x34\x2e\xb1\xb3\x5f\x39\x9f\x99\x8d\xb3\x72\xad\xc3\xde\xce\x43\xc9\x61\x6e\x3e\xa3\x7f\x07\x8e\x47\x33\x0f\xfa\x4e\x1f\x23\x3d\xa0\x61\x91\xfa\xa0\x9f\x23\x67\xe4\x85\x7e\x24\xb3\x73\x1b\x72\x27\xf3\xda\x41\x67\x27\xfd\x2b\xbe\x76\xba\x16\xf6\x48\x3c\x6d\x0e\x2e\xad\x57\x2e\x6d\x3e\x1d\xec\xbc\xa7\xc9\xec\xac\xb9\xd1\x9c\x05\x9b\x05\xe5\xbc\xa5\xd5\x78\xca\x1c\xdd\x2a\xa7\x05\xd3\x4f\x6a\xb9\x3e\x9c\x1f\x81\xe9\xa0\x98\x83\x9b\xe1\x04\x7c\x2d\xdb\x0f\xe7\xb2\xba\x33\x3e\x6a\x34\xce\xf8\x8f\x5f\xcf\xe6\xb2\xe3\x80\xf0\xc7\xe7\xb2\xe3\xab\x32\x97\x3d\xfe\xbf\xf8\xd3\xb9\xec\xf8\xd4\xef\x99\xcb\x9e\x5b\x63\x73\xd9\xff\x02\x00\x00\xff\xff\xe1\x26\x74\xf3\x00\x20\x00\x00")

func templatesGoBytes() ([]byte, error) {
	return bindataRead(
		_templatesGo,
		"templates.go",
	)
}

func templatesGo() (*asset, error) {
	bytes, err := templatesGoBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "templates.go", size: 20480, mode: os.FileMode(420), modTime: time.Unix(1486397048, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testTmpl = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x64\xce\xb1\xce\x83\x30\x0c\x04\xe0\x3d\x4f\x61\xa1\xcc\x41\xac\x28\x64\xfc\xb7\xbf\x53\x5f\xc0\x15\x2e\x41\x22\x10\x11\xb7\x1d\x2c\xbf\x7b\x45\x10\x53\xc7\x93\x3e\xdd\x9d\x08\x53\xca\x0b\x32\x41\xf3\xc0\x42\x6d\x24\x1c\x1b\x70\xaa\xc6\x88\x7c\x66\x8e\x60\xb9\x1f\xc0\xdd\xa9\xb0\xaa\xf1\xb1\x0b\x22\x96\xdd\x0d\x13\xa9\xfa\x36\x76\xc1\xf8\x71\x7e\x07\x03\x00\xe0\x73\xf8\xc3\x79\x79\xed\x54\x7a\xa8\xee\x8a\x87\xcd\x27\x12\xd9\x71\x9d\x08\x6c\x82\x7e\x00\xcb\xee\x9f\x4a\xc1\xa9\x9a\x7c\xb4\xa7\x13\x8b\xd0\x3a\x1e\x9b\x6d\xed\xbf\xa2\xf9\xf9\xfc\xdc\x36\xa6\xbd\xbe\xfe\x06\x00\x00\xff\xff\x16\x62\x6e\x7b\xd0\x00\x00\x00")

func testTmplBytes() ([]byte, error) {
	return bindataRead(
		_testTmpl,
		"test.tmpl",
	)
}

func testTmpl() (*asset, error) {
	bytes, err := testTmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "test.tmpl", size: 208, mode: os.FileMode(420), modTime: time.Unix(1485783225, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"base/footer.tmpl": baseFooterTmpl,
	"base/head.tmpl": baseHeadTmpl,
	"chaos.tmpl": chaosTmpl,
	"index.tmpl": indexTmpl,
	"public/style.css": publicStyleCss,
	"templates.go": templatesGo,
	"test.tmpl": testTmpl,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}
var _bintree = &bintree{nil, map[string]*bintree{
	"base": &bintree{nil, map[string]*bintree{
		"footer.tmpl": &bintree{baseFooterTmpl, map[string]*bintree{}},
		"head.tmpl": &bintree{baseHeadTmpl, map[string]*bintree{}},
	}},
	"chaos.tmpl": &bintree{chaosTmpl, map[string]*bintree{}},
	"index.tmpl": &bintree{indexTmpl, map[string]*bintree{}},
	"public": &bintree{nil, map[string]*bintree{
		"style.css": &bintree{publicStyleCss, map[string]*bintree{}},
	}},
	"templates.go": &bintree{templatesGo, map[string]*bintree{}},
	"test.tmpl": &bintree{testTmpl, map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

