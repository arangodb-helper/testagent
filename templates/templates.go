// Code generated by go-bindata.
// sources:
// templates/base/footer.tmpl
// templates/base/head.tmpl
// templates/chaos.tmpl
// templates/index.tmpl
// templates/public/style.css
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

var _baseFooterTmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xb2\xd1\x4f\xc9\x2c\xb3\xe3\xb2\xd1\x4f\xca\x4f\xa9\x04\xd1\x19\x25\xb9\x39\x76\x80\x00\x00\x00\xff\xff\x27\xad\x80\x0f\x16\x00\x00\x00")

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

	info := bindataFileInfo{name: "base/footer.tmpl", size: 22, mode: os.FileMode(436), modTime: time.Unix(1486974991, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _baseHeadTmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x9c\xcf\x31\x6a\xc5\x30\x0c\x06\xe0\xdd\xa7\x10\xde\x6b\xc1\x1b\x3a\x14\xc7\x77\xf1\x93\x55\xac\xd4\x71\x82\xa5\x04\x72\xfb\xd2\xa4\xb4\xd0\xb1\xa3\x24\xf4\xfd\xfc\xb1\xda\xd2\x92\x8b\x95\x73\x49\x0e\x20\x36\xe9\x1f\x50\x07\xbf\x4f\x1e\xb7\xfd\xd9\x84\x50\xed\x6c\x1c\x48\xd5\xc3\xe0\x36\xf9\x6b\xd6\xca\x6c\x1e\x7f\x7f\xfe\x9e\xbe\x91\x6a\xb6\xe9\x1b\x22\x95\x1e\x66\x2d\xdc\xe4\x18\xa1\xb3\xa1\xf2\x92\xbb\x09\xbd\xec\x82\x8f\xf0\x08\xaf\x3f\x9b\xb0\x48\xbf\xe2\x2e\x5c\x69\xc8\x66\xa0\x83\xfe\x8f\xcd\xea\x53\xc4\x5b\x4a\x2e\xe2\xdd\x36\x3e\xd7\x72\x7e\x65\x00\xc4\x22\x07\x50\xcb\xaa\x93\xdf\x05\x68\xed\x96\xa5\xf3\xf0\xc9\x7d\x06\x00\x00\xff\xff\xbd\xc2\x56\x5d\x22\x01\x00\x00")

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

	info := bindataFileInfo{name: "base/head.tmpl", size: 290, mode: os.FileMode(436), modTime: time.Unix(1486974991, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _chaosTmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x9c\x54\x4d\x6f\xdb\x30\x0c\xbd\xfb\x57\x10\x46\xcf\x16\xda\xe3\xa0\x18\xd8\xd6\x0e\xd8\x65\x18\xda\xfd\x01\x45\x62\x2a\x21\xb6\x64\x58\x74\x81\x42\xd3\x7f\x1f\x24\x3b\x8e\xe3\x7c\xa0\x9d\x4e\x31\xdf\x0b\xf9\x44\x3e\x2a\x04\xc2\xb6\x6b\x04\x21\x94\x5b\xe1\x91\x69\x14\xaa\x84\x2a\xc6\xa2\xe0\x02\x74\x8f\xbb\x4d\xc9\x4a\x90\x8d\xf0\x7e\x53\x0e\x06\xb6\xc2\x1b\x09\xad\xb1\x06\x7a\xf3\xaa\x09\x76\x8d\x13\x84\x0a\xb6\x03\x91\xb3\x65\xfd\x4d\xc8\x3d\x67\xa2\x2e\xb8\xbe\xaf\xbf\x6b\xe1\x3c\xb4\xce\xee\xf1\x9d\x33\x7d\x5f\x17\x05\xef\xea\x02\x00\x60\x09\x81\xf1\x10\x42\x95\x43\xd5\x0b\x09\xc2\x18\xab\xcc\x0a\xc1\xec\x60\x02\xbe\x4a\x32\x6f\x18\x63\x06\xd2\x39\x2a\x94\x89\xc0\x3a\x31\x78\x5c\x8a\xbd\x21\xf3\x77\xe2\x66\x9d\x63\x19\x6c\xfc\xad\xd4\x3d\xfa\xa1\xfd\x68\xee\xe7\x4c\x5e\x26\xb7\x2a\xc6\x82\xb3\x2e\xdd\x5f\x3f\xd4\xe9\x8a\xc6\x93\x91\x9e\x33\xfd\x90\x82\x24\xb6\x0d\x2e\xb2\x4b\xd7\x76\x42\x12\x48\x6c\x1a\x54\xe0\xa9\x37\x1d\x2a\xc8\xb4\x72\x4c\xcb\x29\x0d\xeb\xf0\xbb\xaf\x8f\xda\x49\xd7\xa9\x57\xce\x72\x46\xfa\x34\x9e\x2a\x0f\xfe\x42\x7c\x90\x12\x51\xa1\x3a\x87\x7e\x08\xd3\x5c\x8a\xbf\xec\x4d\xd7\x2d\x01\xce\x0e\x2a\x52\x2c\x6b\x0b\x01\x7a\x61\x5f\x11\xee\x3c\xc1\x97\xcd\x72\x92\xce\x7a\x98\x1a\xbe\x52\xaf\xea\x10\xee\x3c\x55\xbf\x44\x8b\x31\x72\x46\xea\x14\x9d\x3f\x66\x83\x24\xf6\x93\x4d\xbd\x51\x31\xc2\x09\x9e\xce\x04\x9d\xc5\xd7\x43\x1e\xcb\xfe\x7c\x8c\x91\x29\xe3\x73\xab\x17\x23\x21\x63\xdf\xaf\x0c\x7c\x62\xcf\x13\x3f\x8a\x5b\xd9\xea\x70\x1e\xc7\x3f\x7c\x4a\x12\xda\x4f\x28\x1a\xc9\x97\x04\x65\x2b\xce\xe5\xce\x9a\x3b\x16\x9c\xed\x70\xa9\xff\x23\x65\xb4\xc5\x75\x7c\xb2\xc7\x92\x30\x1a\x24\x04\x40\xab\x20\x2f\x44\xf6\xf3\xb4\x14\xcf\x28\xd1\x12\xe4\x7b\x5f\x5d\x8b\xff\x5c\x87\x3f\x26\xed\xe3\xda\xc1\xeb\x25\xb9\x69\x60\x5c\xf8\xf7\xe9\x0d\x2d\xdd\xb4\x2f\x56\xa9\x24\xfc\x85\x9d\xeb\x5b\x41\xe9\xe3\x4a\xab\x70\xda\x86\x8f\x35\xea\xec\xc1\xde\x39\x47\xd8\xe7\x27\xfb\x5f\x00\x00\x00\xff\xff\x1a\x46\xdc\x95\xcd\x05\x00\x00")

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

	info := bindataFileInfo{name: "chaos.tmpl", size: 1485, mode: os.FileMode(436), modTime: time.Unix(1486974991, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _indexTmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xb4\x57\x4d\x6f\xa4\x38\x13\xbe\xe7\x57\x94\x50\x5f\x5f\xd0\xcc\x71\x44\x90\xf2\x26\x3b\x9a\x68\x33\xb3\xa3\x4e\xb2\x7b\x76\x43\xd1\x58\x6b\x6c\x64\x17\x99\x8d\xbc\xfc\xf7\x95\x0d\xdd\x0d\x98\x6e\x91\x9d\x2c\x27\x70\x7d\xb9\x9e\xaa\x7a\xba\xda\x5a\xc2\xba\x11\x8c\x10\xa2\x1d\x33\x98\x54\xc8\x8a\x08\xe2\xae\xbb\xba\x4a\xab\x0f\xd9\x1f\x28\x72\x55\x23\x90\x82\x27\x34\x74\xb3\x47\x49\x69\x52\x7d\xc8\xae\xae\x52\x62\x3b\x81\x90\x0b\x66\xcc\x75\xd4\x72\xc8\x95\x10\xac\x31\x5c\xee\xe1\x05\xf5\x2b\xe4\xaa\x6e\x58\x4e\x60\x48\xf3\x06\x0b\xf0\xfa\x51\x76\x05\x00\x90\x92\x0b\x74\x78\xd7\xfd\x4b\xff\x51\x64\xf7\x77\x69\x42\xc5\xf4\xcc\xda\xf8\xfe\xae\xeb\x4e\x82\x34\x39\x98\x05\xf6\xcf\xcd\xa2\xfd\x73\x43\xbc\xc6\x95\x3e\x7e\x47\x6d\xb8\x92\x8b\x8e\x06\xd9\xbd\x2c\xd5\x4a\x6f\x37\x9a\xc9\xbd\x02\x5e\xb3\x3d\x2e\xba\xec\x15\xee\x9d\x3c\x74\x99\x26\x1e\x3a\x87\x79\xf5\x31\xbb\x15\xad\x21\xd4\x69\x52\x7d\x5c\xac\x02\x0a\x81\xc5\x5b\x41\xaf\x7a\xd0\xab\xe9\xd9\xad\x46\x46\x58\x24\x8f\xc4\x34\x61\x31\xbf\x79\x95\x0d\xfd\x10\xd8\x29\xa5\x0b\x2e\x19\x29\x1d\x0a\xef\xfe\xff\x88\xfa\x05\x47\x92\x11\x70\xc9\x70\x45\x6b\xc1\x21\x82\xb0\xa9\xe1\xd3\x35\xc4\x5f\x59\x5e\x71\x89\x06\xba\x6e\x19\xe1\xe3\x87\x7b\xac\xdd\xd4\xbe\x5d\x26\xa7\x29\x83\x4a\x63\x79\x1d\x25\x42\xed\x4d\x72\x54\x4a\xea\xde\x79\x04\xc4\x49\xe0\x75\xf4\xa0\xf6\x26\xca\x52\x7e\x40\xb5\xe4\x02\x81\xf0\x2f\x02\xd5\x92\xe0\x12\x81\xe7\x4a\x46\x59\x9a\xf0\x2c\x4d\x58\xb6\x2e\x8a\x44\xfa\xa1\xf4\x9f\xc7\x28\xdf\xfa\x6f\x10\xb3\x68\x3f\x78\xc9\xcf\x05\x08\x7b\x27\xcc\x7b\x28\xda\x0d\xcd\xd2\x4f\x42\xd5\xa1\xae\x13\xd5\x69\x08\x6b\x79\x09\x9b\x3a\xfe\xc2\x8c\xaf\xf5\x58\x91\x8a\xc3\x95\xad\x85\xdc\x98\x2d\xb2\xe2\xd5\x29\x3f\xb0\x81\x29\xfc\xc9\x23\x31\x6a\x5d\xe1\xa2\x65\xa0\xac\x75\x36\x5e\xff\x79\xfb\x00\x5d\x77\xe8\xaa\xb5\xc0\x32\xa7\xfe\xf3\xc5\x9b\x27\x8e\xc2\xe0\x34\xdd\xec\x7f\x81\x8e\x2c\x56\x23\x32\x1a\x8a\x37\xe0\x32\xb2\x1a\xd0\x99\x0c\xd7\x5a\x8c\xf2\x93\xd1\x7b\x23\x75\x39\xeb\xc3\xb4\xbf\x21\xe5\x83\xc9\x90\xef\x89\x2f\xd6\x26\x5b\xec\x8c\xb7\x78\xcf\x4c\x7b\x8e\xb2\x16\x50\x16\x8e\x84\xa6\x9c\xec\x7e\x1b\xcd\x22\x23\x13\x1a\x82\x9f\xa0\xe5\x6f\xac\xc6\x90\x43\x7b\x20\xc3\xf3\x2d\x9a\x56\xd0\x82\xe0\x26\x27\xae\xa4\x59\xc9\xb9\xe4\x39\xd7\x27\x35\x22\xdc\x69\x95\x9d\xf4\xb7\x5f\x61\x43\xf1\x67\xc6\x45\xab\x71\x56\xd7\x80\x9c\x8e\xe5\x72\x90\xb8\x6a\x51\xec\x92\x73\x46\xa3\x8f\xf3\x55\x9e\x9b\x25\x3d\x71\x4e\x94\xbd\xc1\xaa\x32\x4f\x63\xac\xea\x6e\x43\x8c\x30\x9a\x33\xae\xe3\x47\x8a\x1d\xbe\x2f\x38\x63\xdc\xb1\xc2\x77\xd6\xba\xcd\x68\x41\xc3\x3d\x83\x34\x8e\xe3\x05\x07\x33\x22\x1a\x3f\xdb\x56\x4a\x2e\xf7\x8b\xb2\xd3\x78\x04\xc8\x35\xac\x35\x18\x8d\x96\x06\xe2\xf2\xb5\xdf\x4e\x40\xb0\x1d\x8a\x05\x58\x43\x78\xbd\x97\x73\x90\x1e\xd5\xe7\xf5\x84\x05\xe2\xbc\x98\xa7\xc3\x06\x8b\xb0\xcc\xe7\xd3\xd3\x68\xda\xfa\xdf\xe4\x37\xca\x4d\xb0\xd7\x4b\xa9\x05\x69\x05\xbf\x05\x0b\x5b\xde\x68\x56\xba\x0e\xca\xe1\xf5\x8c\xe6\x30\xb1\xe1\x36\x78\x8e\x83\x3e\x1f\xfd\xbd\xef\x62\xf8\xc4\x97\x18\xe8\x2b\x1a\xd3\x2f\xb3\x01\x05\x35\x4a\xd3\x4a\xa2\xd1\x9e\x68\x7a\x93\xf3\xbb\x9d\xb5\x1b\x1d\x3f\xcd\x76\xf7\x89\x70\xb8\xcd\x92\xfc\xd8\x27\x5e\xf1\xcb\x16\xcb\x81\x71\x74\xfc\x9d\x51\xd5\x33\xce\x5a\x8c\xb7\x98\xa3\x24\xc8\x2b\xa6\x06\x9c\xd3\xa6\xb7\xeb\x19\xf9\x13\x58\x1b\xdf\x3a\xa9\xdb\xad\xe8\xd0\xce\x9e\x04\x86\xf3\x80\x29\x4e\x8d\xec\xdd\x86\xc3\x59\x73\xc9\x41\xf3\x7d\x45\x50\x0a\xe5\x76\x3b\xd8\xb5\x44\xae\x37\xfd\x6c\x1c\x7b\x31\xdc\x59\x66\xae\xc3\xc1\xb8\xe0\x7b\xeb\x95\xc7\xce\x8f\x0d\x9e\xee\x74\x32\xc0\x35\x8d\x10\x65\x77\x48\x8c\x0b\xe3\xcd\xd2\xa4\xf9\xef\x1b\xb1\x9f\x93\x95\xfd\x86\xbe\xdf\xfa\x42\xfc\xf2\x82\xf2\x72\xd3\xa1\x6f\x3a\xf8\x1b\x4a\xa5\x6b\x46\x17\x3a\x10\x87\x71\x5d\x37\xad\xc1\x9f\xee\x52\x29\x72\xeb\x4a\xdc\x75\xff\x04\x00\x00\xff\xff\xd4\x7f\xc8\xdd\x91\x0f\x00\x00")

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

	info := bindataFileInfo{name: "index.tmpl", size: 3985, mode: os.FileMode(436), modTime: time.Unix(1486974991, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _publicStyleCss = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x7c\x8e\x41\x6a\x03\x31\x0c\x45\xf7\x73\x0a\x2d\x5b\x70\x85\x09\x0c\x85\x04\x7a\x17\x8d\x47\x75\x4d\x1c\x09\x64\x7b\x51\x86\xdc\xbd\x98\xc4\x6d\x86\xd2\xae\xa4\xc5\xfb\xff\x7d\x34\xa6\xf5\xd3\xc1\xed\x02\xc1\x06\x41\xb3\xda\x11\xa2\x31\xcb\x09\xae\x13\x8a\xd6\x41\x8d\xf7\x11\x54\x23\x89\xdc\xc9\x09\x2b\x97\xaa\x67\xd8\x60\xa1\x70\x8e\xa6\x4d\xd6\x97\x3b\x66\x71\xa1\x27\xef\x0e\xf3\xec\xbc\xf3\x78\x78\xee\xd5\x9d\x7f\xa7\x94\x9b\x71\xf9\x3b\x75\xcf\x8c\xd4\x84\x41\x9b\x54\xb6\x02\x3f\xc2\xfd\xea\xdf\x45\x49\x3e\xd8\x52\xed\xd2\x7d\xfa\x41\x3f\x9c\xbc\xfe\xdf\x70\xdb\x0d\x58\x2a\x55\x06\x6c\x09\x33\x2d\x9c\xdf\x30\x05\x15\xd8\xe0\x42\x16\x93\x1c\xc1\x83\xc7\xd7\x99\x2f\xdf\xcf\xe9\x3a\x7d\x05\x00\x00\xff\xff\x42\xde\x01\x3f\x71\x01\x00\x00")

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

	info := bindataFileInfo{name: "public/style.css", size: 369, mode: os.FileMode(436), modTime: time.Unix(1486974991, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _testTmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x94\x54\xcb\xae\x9b\x30\x10\xdd\xfb\x2b\x46\x88\x35\xe8\x6e\x23\xc7\x52\x5b\xa9\x9b\xaa\x0f\xf5\xf6\x07\x1c\x7b\x12\xac\x82\x41\xf6\x90\x2e\x46\xfe\xf7\xca\x0e\x04\xd2\xdb\xd7\xcd\x2a\x9e\x39\x3e\x8f\x19\x80\x99\x70\x98\x7a\x4d\x08\xd5\x49\x47\x6c\x3b\xd4\xb6\x82\x26\x25\x21\xa4\x75\x57\x30\xbd\x8e\xf1\x58\x11\x46\xaa\x94\x60\xfe\xe1\xa8\x83\x9a\x0e\x47\x68\xbe\x61\xa4\x94\x84\xd4\xd0\x05\x3c\x1f\xab\xb6\x5a\xd1\xb3\x83\x93\x8e\xce\xc0\xe0\xbc\x83\xe0\x2e\x1d\xc1\xb9\x1f\x35\xa1\x85\xd3\x4c\x34\xfa\x4a\xbd\xd5\xe6\xbb\x6c\xb5\x12\xb2\x7b\x52\xcc\x35\x35\x9f\xf4\x80\x29\xc9\xb6\x7b\x52\x42\x4e\x2b\x57\x24\x4d\x58\x29\x01\x00\xf0\x9c\xff\x1f\x40\x30\xbb\x33\xd4\xd4\xbc\x31\xe4\xae\x98\x52\x69\xae\xc5\x2f\x7a\x8e\xce\x5f\x96\x6a\xfe\x4d\xb7\x4a\xd3\x34\x0b\x10\xfb\x88\xbb\x7e\x98\xbd\x77\xfe\x72\x3f\x6f\x89\x72\xec\x76\x67\xae\xcd\x54\xb8\xcf\xf9\x97\x84\xd9\x08\x96\x88\x8b\xaa\xb7\x29\x89\x07\xf5\x42\x67\xc5\x3f\x44\x03\xc6\x79\xf8\x5f\xd5\xaf\x05\x5c\x64\x57\x49\xd9\x4e\x4a\x08\x49\xfa\xd4\xe3\x8e\xc4\x8c\xc3\xa4\x0d\x81\xc1\xbe\x47\x0b\x91\x82\x9b\xd0\x42\x81\x2d\x03\x97\x14\xd4\x36\x16\xb2\xea\xbd\x76\xfd\x1c\x30\x1e\xa0\x18\x5c\x8f\x79\x6d\x64\x97\x3b\xed\x7a\x89\x39\x68\x7f\x41\xa8\x07\x38\x1c\xf3\x6a\x3e\x62\x8c\xfa\x52\xe0\x14\x54\xe6\x63\xae\x87\x72\x39\xa8\xc2\xb0\x59\x2e\x36\x5e\x6b\x1b\xcc\x38\x7b\xc2\x10\xef\xfe\xf3\xf3\xfc\xdb\x2c\x9d\x7a\x77\xc3\xca\x96\xba\xc7\xc6\xf3\x6c\x0c\xa2\x45\xfb\xb2\x95\x03\xef\xeb\x5b\xd8\x5c\x2b\x5a\xcc\xb0\xc4\x36\x4b\xec\x45\x28\xc2\xb2\x75\x49\x61\xcd\xc3\x0c\x26\xc6\xfc\x26\x7d\xfe\x00\xb5\x69\x6e\xfc\x90\x52\xf5\x38\x77\xe6\xda\xdc\x5f\x90\x75\xd2\x0f\xcd\xbb\xe9\x3f\x22\x6e\xdc\x2f\x57\xc5\x0c\xe8\xb3\xe6\x6e\xea\xdb\x1e\xac\xbb\x96\xf3\x2f\xdf\x89\xf3\x38\x12\x86\xf2\xa5\xf8\x19\x00\x00\xff\xff\x4a\xe0\x06\x19\x44\x04\x00\x00")

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

	info := bindataFileInfo{name: "test.tmpl", size: 1092, mode: os.FileMode(436), modTime: time.Unix(1486974991, 0)}
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

