// Code generated by go-bindata.
// sources:
// templates/base/footer.tmpl
// templates/base/head.tmpl
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

var _indexTmpl = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\xb4\x96\xdf\x4f\xac\x3a\x10\xc7\xdf\xf9\x2b\x26\x64\x5f\x2f\x8d\x3e\x1a\x96\xc4\xeb\x5e\xa3\xb9\x7a\xaf\x59\x35\xe7\xb9\x0b\xb3\x4b\x73\x80\x92\x76\x30\xc7\x34\xfd\xdf\x4f\x5a\x60\x05\x41\x0f\x26\xab\x4f\x38\x3f\xda\xe9\x67\xa6\xdf\xad\x31\x84\x65\x5d\x70\x42\x08\x77\x5c\x23\xcb\x91\x67\x21\x44\xd6\x06\x41\x9c\x9f\x25\x3f\xb0\x48\x65\x89\x40\x12\x9e\x50\xd3\xe5\x01\x2b\x8a\x59\x7e\x96\x04\x71\x9d\x04\xb7\x9b\x0b\x30\x26\xba\xdd\x58\x1b\xc4\x3b\xc5\x92\xe0\xb9\xf6\x96\xe7\x9a\x44\x89\xce\xca\xea\xc4\xad\x74\x9e\x5c\x15\x8d\x26\x54\x31\xcb\xcf\x9d\x85\xf8\xae\x40\x48\x0b\xae\xf5\x3a\x6c\x04\xa4\x58\x14\x98\x81\x26\x25\x6a\xcc\xc0\xbb\xc3\x24\x00\x00\x88\xc9\x15\xd5\x7f\xab\xf6\xa3\x73\x24\xb7\x9b\x98\x51\x3e\xb6\x5d\x29\xe4\x84\x19\x7b\x24\xae\x08\xb3\x98\x51\x36\x0e\xe8\x8e\x31\xc9\x93\x52\x65\xa2\xe2\x24\xd5\xd4\xb9\xf9\xfb\x11\xd5\x0b\x0e\x3c\x31\xeb\x8b\x71\x36\x5f\xa2\x31\xa0\x78\x75\x40\x58\x95\x70\xb1\x86\xe8\x9e\xa7\xb9\xa8\x50\x83\xb5\x73\xe5\x0f\xca\x72\x7f\xc6\xac\xca\x16\xe6\xd0\x1a\x73\xc8\x15\xee\xd7\x21\x2b\xe4\x41\xb3\x63\x10\x2b\xdb\xc5\x43\x20\x41\x05\xae\xc3\x3b\x79\xd0\x61\x12\x8b\x9e\xea\x5e\x14\x08\x84\xbf\x08\x64\x43\x85\xa8\x10\x44\x2a\xab\x30\x89\x99\x48\x62\xc6\x07\x85\xbc\x03\x34\x57\x56\xc7\xf4\x92\xde\x55\xc7\xa6\xa1\x1d\xf6\x51\xe8\x78\x0b\x63\xc4\x1e\x56\x65\x74\xc3\xb5\x6f\xc5\x30\x90\xb2\xbe\x7e\x63\x20\xd5\x7a\x8b\x3c\x7b\x75\xc1\x77\xbc\x9b\x3f\x6f\x79\x24\x4e\x8d\xe3\x1a\x26\xb3\xb4\x8c\x71\x39\x3e\xfe\x79\x7b\x07\xd6\xf6\x4d\xe7\xf3\xf1\x13\xba\xdc\x85\x9f\x9a\xad\x31\x58\x68\x1c\x1f\x37\xf9\x6b\x12\x53\x65\x8b\x89\x0c\x66\xf6\x0b\x5c\x06\x59\x1d\x9d\xd1\xec\x2f\x65\x94\xbe\x25\x7d\xc3\x14\x7e\x72\xea\xfe\x32\x7e\xe1\xc8\x7d\x4a\x77\xde\xb7\xeb\xbc\xf4\xb0\xd9\x4e\xfb\x8c\x53\x9e\xb4\x95\x10\x63\x00\xab\x0c\xbc\x5c\x7a\xe1\xeb\x24\xd3\x29\xae\x3e\xb5\x60\xfe\xc7\x4b\x9c\xaa\x9b\x63\x88\x0b\xa5\x8d\xbc\xb4\xf9\xe2\x3e\xd4\xb5\x23\x45\x42\x4d\x0e\x22\x45\x6e\x63\xd7\xa3\xc1\x3f\x7f\x68\xb9\xdb\xe2\xff\x7f\x61\x45\xd1\x35\x17\x45\xa3\xb0\x6d\xb2\x5f\xa0\xb7\x58\x0b\xfb\xee\x73\x29\xd6\xeb\x63\xfc\x69\xc9\x3e\x89\x39\xb2\xf7\xa8\x35\x3f\xcc\x38\xb6\x58\x4b\x45\x0b\x99\x2b\xcf\xbc\x4d\xf9\x98\xba\x31\x2b\x15\x3d\x89\x59\xb4\xad\xb3\xab\x66\xce\x7f\x6c\x99\x0f\xbc\xd9\xe2\xbe\x83\xad\xa2\x07\x4e\xb9\x4b\xe1\xc9\x52\xc6\x5b\x4c\xb1\x22\x48\x73\x2e\x3b\xce\xee\xbd\xe0\xf2\xae\x9c\x09\x4a\x59\xfd\xc4\x57\x10\xda\xbd\x16\xbc\x29\xf2\x23\x68\x6d\x14\x40\xff\x0b\xd1\x39\x2e\x53\x12\x2f\x23\xd1\x3c\x5e\x51\xbf\x3e\xab\x79\xa3\x31\x1c\x74\xb0\x14\x95\x00\x25\x0e\x39\xc1\xbe\x90\xee\x77\x0b\x76\x0d\x91\xbb\x8c\x0f\x2e\xf6\x78\x1b\xa7\x7a\xfc\x6e\x69\x85\xba\x29\x97\xae\xbd\xf5\xc1\xc3\xc5\xbd\x90\xfb\x37\xd0\x37\x4f\x99\x63\x24\xab\x85\xc3\x84\x7e\x98\x5a\xb8\xff\xbc\x60\xf5\xf9\x44\xe1\x67\x13\x85\x51\xbb\xf3\xd0\xfd\xf1\x64\x4c\x5e\x9b\x7b\x29\xc9\x29\x6a\x64\xed\xef\x00\x00\x00\xff\xff\xbb\xe8\xe2\x95\x8a\x0a\x00\x00")

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

	info := bindataFileInfo{name: "index.tmpl", size: 2698, mode: os.FileMode(420), modTime: time.Unix(1485877353, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _publicStyleCss = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\x6c\xcc\xc1\x09\x43\x21\x10\x84\xe1\xbb\x55\x4c\x01\x89\x05\x24\xd5\x6c\x74\x23\x41\xd9\x85\x55\x0f\x41\xec\x3d\x48\x10\x3c\xbc\xd3\xcc\xe1\xe3\xf7\xc6\x14\xbf\x37\xfc\x17\x84\x81\xa0\x45\xed\x81\x64\xcc\xf2\xc4\x74\x5e\xb4\x6d\xb5\xef\x09\xd5\x48\x12\x2f\xe9\x7c\xe3\xda\x34\x63\xe0\x45\x21\x27\xd3\x2e\xf1\x7e\xf6\x56\x6e\x99\x37\x7d\x4a\x37\xae\x97\xd2\x38\x62\xba\x5f\x00\x00\x00\xff\xff\x5d\xf4\x8d\x91\x9a\x00\x00\x00")

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

	info := bindataFileInfo{name: "public/style.css", size: 154, mode: os.FileMode(420), modTime: time.Unix(1485783225, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _templatesGo = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\xc4\x56\x5d\x8f\xdb\xd6\x11\x7d\x26\x7f\xc5\x8d\x80\x04\x52\xa1\x4a\xfc\xfe\x10\xe0\x97\xd8\x0e\xea\x87\x3a\x40\xeb\x3e\x75\x8a\xe0\x5e\xf2\xde\x0d\x51\x49\x54\x49\x2a\x9d\x5d\x63\xff\x7b\x71\x38\x57\xb2\x76\xbd\x89\x0b\x23\x40\x1e\x24\x91\xf7\x63\xe6\xcc\xcc\x99\x33\xda\x6e\xd5\xeb\xbe\xb5\xea\xce\x1e\xed\xa0\x27\xdb\x2a\x73\xaf\xee\xfa\x3f\x9b\xee\xd8\xea\x49\x6f\xc2\xed\x56\x8d\xfd\x79\x68\xec\xb8\xc3\xf3\x64\x0f\xa7\xbd\x9e\xec\xb8\x35\x7a\xb4\x5b\xd7\xf7\x93\x1d\x36\xd3\xe1\xb4\x7f\x61\xf7\x67\xab\xdb\x17\xf6\xba\x63\x6b\xf9\x85\xf5\xd3\xd9\xec\xbb\x66\x3b\x4e\xf7\x7b\xbb\x69\xc6\xf1\xe9\xee\xf5\x69\x73\xd7\x3f\xdf\x19\xa7\xab\xb9\x37\x3f\xaa\xf7\x3f\x7e\x50\x6f\xdf\xbc\xfb\xf0\x4d\x18\x9e\x74\xf3\x6f\x7d\x67\x3f\x1d\x0e\xc3\xee\x70\xea\x87\x49\x2d\xc3\x60\x61\xee\x27\x3b\x2e\xc2\x60\xd1\xf4\x87\xd3\x60\xc7\x71\x7b\xf7\xd0\x9d\xb0\xe0\x0e\x13\x7e\xba\x5e\xbe\xb7\x5d\x7f\x9e\xba\x3d\x5e\xfa\xf9\xc2\x49\x4f\x3f\x6f\x5d\xb7\xb7\x78\xc0\xc2\x38\x0d\xdd\xf1\x6e\xde\x9b\xba\x83\x5d\x84\xab\x30\x74\xe7\x63\xa3\x7c\x22\xff\x66\x75\xbb\xc4\x83\xfa\xe7\xbf\xe0\x76\xad\x8e\xfa\x60\x95\x5c\x5b\xa9\xe5\x65\xd5\x0e\x43\x3f\xac\xd4\xc7\x30\xb8\x7b\x98\xdf\xd4\xee\x95\x02\xaa\xcd\x7b\xfb\x5f\x18\xb1\xc3\x72\x86\x8d\xf7\xef\xcf\xce\xd9\x61\x36\xbb\x5a\x85\x41\xe7\xe6\x0b\xdf\xbc\x52\xc7\x6e\x0f\x13\xc1\x60\xa7\xf3\x70\xc4\xeb\x5a\xb9\xc3\xb4\x79\x0b\xeb\x6e\xb9\x80\x21\xf5\xed\x7f\x76\xea\xdb\x5f\x16\x82\x64\xf6\xb5\x0a\x83\xc7\x30\x0c\x7e\xd1\x83\x32\x67\xa7\xc4\x8f\x38\x09\x83\x9f\x04\xce\x2b\xd5\xf5\x9b\xd7\xfd\xe9\x7e\xf9\x9d\x39\xbb\xb5\xba\x7b\x58\x85\x41\xb3\x7f\x7b\x41\xba\x79\xbd\xef\x47\xbb\x5c\x85\xbf\x17\x1e\x98\x11\xfb\xbf\x62\xc8\x0e\x83\xe0\xf6\x8b\xe6\xec\x36\xdf\x03\xfa\x72\xb5\xc6\x89\xf0\x31\x0c\xa7\xfb\x93\x55\x7a\x1c\xed\x84\x94\x9f\x9b\x09\x56\xe6\xf8\x7c\x3d\xc2\xa0\x3b\xba\x5e\xa9\x7e\xdc\xfc\xd0\xed\xed\xbb\xa3\xeb\xaf\xf7\x7c\x09\x2f\xeb\x37\x16\xe6\x1a\x2a\xe5\xcb\x18\x06\x63\xf7\x30\xbf\x77\xc7\xa9\xc8\xc2\xe0\x80\xce\x52\x57\xa3\x7f\xed\x5b\x3b\x2f\x7e\xe8\x0e\x56\x81\x26\x1b\x3c\xc1\xcf\x4c\x95\xa5\xeb\x9e\xfb\x5a\xa9\xf7\xfa\x60\x97\x2b\xef\x01\x3e\x7d\x94\xae\xdb\xc0\x7b\xf8\xf8\x1b\x77\xff\xde\x3d\xe0\xee\x8c\xe6\xe9\x55\x00\xfd\xcd\xab\xc0\xba\x5c\xdd\x22\x7f\x6a\x00\xa1\x7d\xc9\x00\x82\x5b\xae\x3e\x05\xfa\x99\x05\x1f\xfd\xaf\x1b\x79\x37\xbe\xe9\x86\xe5\x4a\x99\xbe\xdf\xdf\xde\xd6\xfb\xf1\x0b\x91\xdf\x8f\x12\xb8\x1d\x9c\x6e\xec\xc7\xc7\x9b\xdb\x9e\x12\x60\xf9\x4f\x90\xa9\x1f\x66\x0d\xfb\x70\x38\xed\xd5\x2b\xcf\x86\xe5\x82\x38\x76\xc4\x95\x21\x8e\x2a\xe2\x28\xf2\x9f\x9a\xb8\xb0\xc4\x95\x5f\x73\x8e\xd8\x24\xc4\x6d\x4c\x9c\x39\xe2\xa6\x26\x4e\x1a\x62\x93\x12\xdb\xf4\xd9\x9e\x96\x5f\x5d\x13\x47\x99\xac\xc7\x38\x9f\x13\x9b\x9a\x38\xad\x89\xcb\x82\xb8\x8a\x6e\xfc\x79\x1f\xf8\x24\x25\xb1\x6e\xfd\xbe\x23\x8e\x8b\xdb\x73\x8b\xab\xde\x3c\x89\xc8\xf7\xc1\x4b\xfa\x72\xe9\x96\x1b\x7d\x0a\x83\xe0\x59\x46\xd6\x61\x10\x2c\x9e\x2b\xfd\x62\x1d\x06\xab\x2b\x6b\x9f\x5e\x80\xaf\x3f\xcd\x8d\x76\xeb\x6b\xee\xb4\xab\x9c\xbd\x08\xf1\x4b\x72\x71\xed\xf2\xb9\x4f\x61\xe5\x69\xcd\x3f\xa2\x1b\x76\xea\x05\xb0\x0a\x64\xdf\xa9\x24\x59\x2b\xb0\x76\x77\x4b\xea\x65\x96\x44\xab\x79\x1d\x5c\xdc\x09\x57\xff\x71\xec\x78\x19\x67\x55\x5e\x56\x69\x92\xe4\x6b\x15\xad\x1e\xc3\x40\xc3\xe9\x77\x73\x68\x1f\xe7\x78\x76\xca\x87\x05\x44\xbb\xf9\xfb\xf1\x9a\x56\xbd\xfe\x8c\x67\x7f\xb1\xba\xfd\x6a\x96\xd5\x0d\x71\xe3\x88\xd3\x98\xb8\xd0\xc4\x4d\x4e\x9c\xe2\x6c\x43\x1c\x15\xc4\x36\x22\x6e\x5b\x62\x5d\x12\xc7\x78\xb6\xc4\x85\x21\x6e\xc0\x32\x43\x9c\x6a\xe2\x38\x23\x6e\x4a\xe2\xb2\x24\x76\x31\x71\x9d\x12\xe7\x39\xb1\x6e\x88\xdb\x8c\xb8\x8c\x89\xab\x84\x58\xe7\xc2\xd0\x32\x21\x76\x86\xb8\xc5\x5a\x46\x6c\xc0\xda\x88\xd8\xc4\xc4\x3a\x25\x4e\x32\x62\x87\x4f\x4b\xec\x1a\x59\x6f\xb5\x9c\xaf\x13\x89\xab\xce\x89\xcb\x94\x38\x03\xeb\x2d\x71\x12\x11\xa7\xc0\x5b\x4b\xec\x39\x62\x28\x89\x8d\x93\xee\x88\x2d\xb1\x29\xc5\x66\x5b\x13\x57\x99\x9c\xb1\x2d\x71\xd1\x10\xc7\x0d\x71\x56\x11\xb7\x39\x71\x93\x4a\xdc\xb0\xe7\x6a\x89\xb7\x2d\xa4\xd3\xe6\xb3\x96\xb8\x74\xf2\x71\x96\xb8\x86\x6d\xfc\xfa\x1c\x9a\x0b\x0e\x43\x9c\x24\x82\x15\x77\x8a\x82\x38\x81\xff\x86\xd8\x66\xc4\x71\x45\xac\x63\xe9\x6a\xfc\x3a\x1f\x9f\x31\x52\x2b\xd3\x12\xdb\x46\x72\x57\xc1\x57\x24\xf5\xd4\xa8\x17\x6a\x67\x88\x4d\x24\xb8\x11\xa7\x4d\x88\x13\x4b\x9c\x37\xc4\x45\x4d\xdc\x54\xe2\x53\x47\xc4\x55\x2a\x58\x61\xa7\x81\x5d\x4d\x9c\xa7\xc4\x0d\xf2\x60\x88\x33\x2d\x77\x61\x03\xf5\x46\xec\xa9\x25\x6e\x4b\xa9\x57\x09\xfc\xb9\xf0\x06\x77\x10\x17\xf2\x8b\x1c\x36\x86\x58\x6b\xa9\x7b\xeb\x88\xa3\x9c\xb8\xa8\x24\xb7\x75\x21\x75\x77\xa9\xe0\x87\x92\x95\xad\x70\xeb\xb9\x12\x21\xde\x26\x21\xce\x0b\xe2\xbc\xf5\x3e\xe2\x17\x95\xe8\xc2\xf9\xaf\xd3\xa1\xcb\xed\x4f\x2a\x74\xfd\x47\xf9\xb9\x06\x5d\x0e\xff\xbf\x0a\xf4\x0c\xda\xef\xaa\x3f\x37\x30\x2f\xea\x53\x47\x7f\xa0\xfc\xcc\xff\xb8\xbf\x7e\xc2\x65\xc2\x8f\xd6\xf7\x28\xf4\x62\xd6\x93\x48\xf4\x04\xeb\xe8\xbf\x04\x7d\x54\x10\x17\xe0\x2a\x26\x16\xec\xb7\xc2\xcf\x58\x8b\x0d\x70\xd2\x1a\xe2\xdc\x8a\x86\x60\xf2\x95\x5a\x7a\x25\xc7\x14\xcc\x89\x6d\x29\xeb\x91\x91\x9e\xcb\x8c\x68\x08\x26\x1f\xfa\x0e\x53\x32\xf5\xbe\x53\xaf\x3f\x17\x6c\x39\x7a\x3f\x12\x6e\x67\xb1\x4c\x4a\x60\xd2\x46\xf6\xd3\x4a\x7a\x12\x3a\x85\xde\x2f\x30\x55\x7d\x6c\x98\xae\xd0\x57\x68\x0e\x7a\x08\xdc\x2e\xd1\xb7\x89\xe4\x08\xba\x89\x9e\x4d\x52\xe9\xa5\x59\x47\x4a\xe2\x04\x13\x3e\x13\xed\x41\xbf\xc0\x2f\xf4\xba\x76\x32\xdd\xe1\xef\xd2\xff\xb0\x5b\xd5\xc4\x59\x44\x1c\x27\xa2\x49\xe8\xcd\x16\x5a\x96\x4b\x1f\x21\x8f\x95\x96\x7c\xa0\x9f\x91\xb7\xc8\xeb\x73\xdd\xca\x33\x34\x12\xba\x82\x1c\x21\x17\x88\xd1\x68\xd1\x84\xbc\x12\x4d\x43\xae\x53\x23\xb3\x02\x79\x83\x56\x22\xaf\xae\x10\x9b\xc8\x73\xad\x05\x3b\x30\x35\x56\xb4\x13\x9a\x83\xbb\xc8\x47\x99\x79\x8c\xd0\xd3\x5c\xea\x09\x5f\x79\x26\xb9\x42\x1e\x30\x97\x80\xab\x82\x7e\x40\xf7\xac\xc4\x87\x78\x81\x0f\xef\x2e\x17\xce\x40\x97\x81\x1f\xf3\x64\xc6\x09\xfb\x15\x71\x15\x8b\x5d\xe4\x6b\x9e\x6f\x98\x77\x91\x68\x6c\xe3\x67\xcd\x85\xa3\xe0\x26\xf4\x09\x78\x30\x07\xa0\xef\xa8\x2f\xe2\x02\xa7\x4a\x7f\xfe\x92\x23\xd8\xcf\x63\xe1\xa1\xb9\x68\x56\x2d\xda\x8e\x5a\xe3\x5f\x58\x69\xe4\x0e\xb0\xa2\xd6\xe0\xdd\xf5\x5e\x21\xb3\xcb\x02\xa7\xf5\x71\xd7\x7e\x7e\x26\x52\x53\x9d\x78\x5d\xcf\x04\x13\xe6\xe9\x9c\xeb\x5a\x66\x68\x1c\xcb\x3a\xf2\x01\x5b\x69\x23\x76\xd0\x0b\xe8\xc1\xd2\xfb\x41\xdf\xc0\x16\xf6\x30\x0f\x4a\xf0\x2e\x97\x7c\x22\x5f\xf3\x3c\xf5\x33\xb8\x2a\x04\x13\xea\xa3\x7d\xcf\xe8\x4a\xf8\x84\x39\x62\xfc\xec\x9d\x79\xe5\xff\x81\xce\x33\x02\xcf\x85\xf4\xf1\xff\x02\x00\x00\xff\xff\x29\xa3\x45\x0d\x00\x10\x00\x00")

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

	info := bindataFileInfo{name: "templates.go", size: 12288, mode: os.FileMode(420), modTime: time.Unix(1485877358, 0)}
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

