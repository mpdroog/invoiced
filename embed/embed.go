// Code generated by go-bindata.
// sources:
// embed/invoice.xml
// DO NOT EDIT!

package embed

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

var _embedInvoiceXml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x58\x5b\x6f\xdb\x38\x13\x7d\xf7\xaf\xd0\x27\x04\xe8\xcb\x27\x59\xb2\x13\xb7\x11\x12\x75\x7d\x2b\x60\xc0\x49\xbb\x89\x53\x2c\xfa\x46\x53\x53\x85\x58\x89\x34\x44\xca\xb1\x37\xc8\x7f\x5f\x50\x17\x8a\xba\xd8\x89\xb3\x05\x76\xf3\x10\x20\xd4\xcc\x99\xc3\x99\xa3\x43\x2a\x57\x9f\x77\x71\x64\x6c\x21\xe1\x84\xd1\x6b\xd3\xb5\x1d\xd3\x00\x8a\x59\x40\x68\x78\x6d\x3e\xac\xbe\x58\x9f\xcc\xcf\x7e\xef\x6a\x41\xb7\x8c\x60\x30\x76\x71\x44\xf9\xb5\x99\x26\xd4\x63\x88\x13\xee\x51\x14\x03\xf7\xf8\x06\x30\xf9\x49\x30\x12\x84\x51\x2f\x5d\x47\x1e\xc7\x8f\x10\x23\x6f\xc7\x03\xaf\xc8\xb5\x06\x66\xcf\x28\x7f\x32\x1c\x0f\x23\x7c\x22\xd6\x94\xc5\x31\xa3\xe3\x30\x4c\x20\x44\x02\xa6\x2c\xde\x30\x0a\x54\xf0\x2e\xf4\xf5\xfb\xd0\x27\x88\x13\x7c\x14\x79\xc7\xc9\xb5\xf9\x28\xc4\xc6\xeb\xf7\x9f\x9e\x9e\xec\xa7\xa1\xcd\x92\xb0\x3f\x70\x1c\xb7\xff\xc7\xcd\xf2\x3e\x43\xb4\x08\xe5\x02\x51\x0c\x7a\x3a\x27\x45\xbd\x25\xcb\x29\xbc\xb7\x97\x46\x51\x3e\x60\x98\xdb\x59\xba\xc5\x36\x40\x33\x1e\xe9\x3a\xea\x33\x6e\x3d\x4c\x96\xd6\xc0\x76\xfb\x3b\x1e\xf4\x63\x44\x68\xc0\x70\x5f\xae\x29\x0c\xdb\xb5\x77\x3c\x30\xfd\x5e\xcf\x30\xae\xf0\x1a\x7b\xd3\x94\x0b\x16\x93\xbf\xb2\xd2\x8b\x99\x2f\xa9\x61\xa0\x36\xa4\x1e\x50\x77\x74\x39\x74\xbd\x81\xe3\x7e\xbc\xea\x77\x05\x2b\x94\x6f\x09\xfb\x49\x22\x28\xf2\x65\x7f\x30\xd0\x35\x21\x12\x66\x93\x3f\xf3\xd6\x84\x38\xe7\xde\x16\x92\x81\xed\xe4\x70\x55\x96\x02\x5a\xcc\xfc\xe7\xe7\x47\x11\x47\x86\x7d\x03\x02\xd9\x05\x71\x12\xbc\xbc\xe4\x39\x32\xb8\x8c\xe5\x3c\x85\x19\x12\xd0\x48\x91\xcb\x01\x12\xa0\x52\x54\x5c\x99\x39\xeb\xca\x9b\xd5\xb3\x66\x55\x8e\x2a\x98\x93\x59\xed\x37\x30\x65\x01\x18\x11\xe1\x62\x31\xbb\x36\x1f\x6e\xa7\x4b\xd7\x71\x5c\x33\x5b\x19\x87\x40\xf1\x5e\xae\x8f\x4c\x7f\xf8\xa9\xd8\x6b\x23\xb7\xa2\xc2\x70\x1a\x03\x15\xd3\x34\x49\x64\x5e\x0d\x78\x71\xff\xd5\x38\x1f\xb8\x1f\x8d\x71\xb4\x79\x44\x1d\xf0\xf3\x87\xbb\x82\x6d\x07\x4c\x56\xe3\x7f\x96\x65\xac\xbe\xce\xbe\xda\xb6\x81\x91\xa2\xf1\x0d\x12\xc2\x02\x3f\x53\x69\x46\xe3\x5e\xa0\x44\x64\xfb\x95\xd3\xb6\x5c\xc7\x72\xdc\x1c\xb9\x7a\x52\x45\xcf\x69\x50\x8b\x1d\x16\xb1\xe5\xba\x2c\xdc\x6f\x95\xb3\xac\x62\xce\x08\x7b\x63\x8c\x59\x4a\x05\xa1\xe1\x7d\xba\xd9\x44\x04\x92\x6f\x28\x11\xfb\xb2\x06\xc2\x9e\xf6\xb7\xbe\x72\x8b\x62\xf0\xd5\xdb\x95\xb1\xc9\x96\xca\x51\xce\xa9\x20\x62\x6f\xcb\xb5\x72\x92\x7a\x4a\x4e\xab\x85\x94\xe3\x33\x2e\x50\x34\x0e\x82\x04\x38\xd7\x6b\x48\xbe\xf9\xea\x92\x50\xad\x7a\x49\x20\x5b\x6d\x10\xb8\x17\x09\x80\x70\x4b\x0e\xa7\x26\x0e\xba\x13\x73\xfa\x9d\x64\x32\x9a\x53\xd9\xd4\x64\xaf\x57\xca\x95\x1b\x00\x15\xca\x61\x9a\x1a\x1b\xba\xa3\x91\xe5\x7a\x99\xc8\x06\x1d\x2a\xbb\x5d\x16\x1a\x6e\xa1\x34\x99\x35\xea\x17\xdd\xae\xf7\xb5\x35\xd2\x15\xda\x65\xf6\xd9\x9c\xab\xf4\x63\x44\xf7\x8b\x99\x91\xb9\x21\x48\x36\xb7\x4b\xef\xfb\x78\x65\x16\x2b\x15\xcb\x1f\x3f\x7e\x98\xaa\x93\x13\x44\xff\xb4\xbf\x23\x51\xf6\x50\x01\x35\xba\xd5\x51\x58\xd9\x90\x56\xf3\xe1\xb6\x3f\x9f\xce\x8d\x0b\xf7\x62\xd8\x2e\x3c\x32\xfd\xef\xe3\x95\xee\x4d\xfa\xd6\x5b\x15\x34\xfd\xb5\x9f\xa9\x47\x4b\x08\x51\x94\xeb\xa1\xd1\x93\x3b\x08\x09\x17\x49\x36\x80\xd7\x74\xdf\x8a\x3d\xd0\xde\x7a\xdf\xa6\x0c\x1f\xec\x9b\xc6\xbe\x83\x62\xa1\x40\x2a\x10\x16\x8d\x5a\x2b\x88\x60\xf3\xc8\x28\xf8\xce\xc8\x72\x87\x97\x97\xce\xe5\xe5\xa7\xbc\x48\xf5\xa8\x9e\x32\x8f\x00\x8b\x84\x51\x82\x6f\x10\x89\xfc\x84\x31\x11\xc0\xf6\xb7\x30\x46\x24\xb2\x31\x8b\x0b\xc7\xa9\x47\xf5\xea\x62\xd4\xa8\x68\xd4\x2b\x7b\x3a\xe8\x42\x6d\x9b\xca\x8f\xbe\x5f\x66\x53\x25\xdc\xbf\x68\x54\x8a\xc2\xe9\x56\xd5\x48\x3d\xd5\xac\x3a\x6d\x21\x7f\xf4\xb6\x57\xe1\x95\x77\xa1\xb3\xb9\x87\xde\x86\x63\x9a\x7e\x45\x34\x0d\x4d\xf4\x14\xef\xbd\x3c\x8a\x6f\x00\x51\xae\x1d\x9a\xfa\x72\xeb\xfa\x70\x7e\x3e\x72\x4d\xbf\x3c\x47\x9b\xa1\x35\xbd\xed\x01\xbe\x10\x8a\x28\x26\x28\x2a\xb8\x54\xea\x68\x78\xd7\x87\xc5\x64\x7c\xfb\xa1\xfe\x82\x2f\xd6\x88\xd6\x6f\x53\x0a\x5c\xe1\x2e\x28\x17\x44\xa4\xb2\x57\x93\x04\x51\xfc\xd8\x50\x59\x57\xe0\x51\x13\xfd\x30\x59\x4c\x1b\x3c\x26\x04\xb7\x69\x94\x4d\x3e\x56\xe0\x70\x84\xce\xb5\x9c\xdd\x81\x76\xa9\xc7\xda\xa8\x7a\xd5\xb9\xb0\x62\x02\x45\xda\xf0\x56\x68\x37\x8e\x65\xae\x81\x8b\x1b\x96\x9c\xdc\xfc\xe1\xae\x3a\x75\xb2\x14\x7b\x85\x76\xe5\xae\x54\x8e\x36\x3e\x69\xfa\xe9\x5a\x54\xe8\x15\x3e\x5a\x47\xf0\xb6\x1a\x73\xbd\x44\x95\xd6\x04\xfc\xa7\x84\x2b\xca\x53\x24\x20\x64\x49\xf3\x30\x6a\x9c\x92\xd3\xe5\xc5\xd0\xb9\x30\xfd\xfb\x8e\xa1\x66\xaa\x86\x04\x03\x15\xfe\xc0\x55\x1f\x00\xc5\xca\x7f\xe4\x58\x6e\xed\xb4\x8a\xd6\x66\xa6\x16\x0b\x8d\x94\xaa\xc9\xdc\xe3\x86\x51\x10\x28\xd9\x37\xf5\x23\xed\x6f\xbe\x13\x40\xe5\x47\xf7\xa9\x53\xee\x48\xae\x4b\x73\xbe\xc3\x51\xca\xc9\xf6\x3d\xfa\x69\xe4\xd6\x81\x17\xf4\x24\xe0\xec\xb7\x86\xdd\x48\xd7\xcd\x30\x81\x0d\x22\xc1\x21\x58\xa7\xfa\x44\xd4\xe2\xea\x66\x2a\x85\x7f\xc7\x52\x1a\x10\x1a\x1e\xc4\x51\x7e\xda\x8e\x6e\xa3\xbd\x63\x93\xb5\xcc\x4a\x1d\x5d\x5a\xe8\x19\xc6\xf3\x73\x82\x68\x08\xc6\x19\xf9\xbf\x71\xb6\x35\xbc\x6b\xc3\x96\xb3\xe5\x2f\x2f\x95\x43\x14\x9f\x4c\xb5\xf3\x52\x7d\x19\x9f\x91\x2e\xe7\xae\x3e\x2f\x83\xdf\x53\x94\x9d\x5f\x46\x4a\x89\x90\x47\x87\xbc\x1a\xab\x1d\x9c\x6d\xed\x32\x40\xe1\x34\x12\x6b\xa8\xa7\xa8\xf6\x6c\x5b\xef\xcc\x41\xd1\x96\xfb\x14\x10\x1f\xb9\x21\x9d\x6d\xed\x19\x70\x9c\x90\x8d\x34\xf6\xae\x1b\x52\x79\xe1\x8c\x10\xe7\xe4\x27\x81\xa0\xd3\xa8\x4e\xb6\xaa\xb7\x9b\xd5\x31\xbb\xfa\x65\x86\x75\xd0\xb2\xd4\x2d\xf7\xd8\xfe\x8b\xaf\x70\xad\xd5\xf9\x25\x22\x21\xb8\x79\x3d\xcd\xd6\x5e\x9f\x71\x16\xa6\xd4\x5f\xe5\x34\xd0\x26\x88\xc3\x01\x2d\xba\xb6\x53\x74\x54\x0f\x6a\x5c\xc7\x2a\x82\xb5\x7f\x24\x94\x6f\xc5\xf3\x33\xd0\xe0\xe5\xc5\x30\x7a\x57\xfd\xe2\x91\xff\x77\x00\x00\x00\xff\xff\x57\x1b\x32\x10\xd8\x14\x00\x00")

func embedInvoiceXmlBytes() ([]byte, error) {
	return bindataRead(
		_embedInvoiceXml,
		"embed/invoice.xml",
	)
}

func embedInvoiceXml() (*asset, error) {
	bytes, err := embedInvoiceXmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "embed/invoice.xml", size: 5336, mode: os.FileMode(420), modTime: time.Unix(1528300192, 0)}
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
	"embed/invoice.xml": embedInvoiceXml,
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
	"embed": &bintree{nil, map[string]*bintree{
		"invoice.xml": &bintree{embedInvoiceXml, map[string]*bintree{}},
	}},
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
