package file

import (
	"bytes"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	dir := "./test-home"
	os.MkdirAll(dir, 0755)
	defer os.RemoveAll(dir)
	os.Setenv("HOME", dir)
	var in bytes.Buffer
	in.Write([]byte("account_id\napi_key\nY\n"))
	Create("hello, world")
	assert.FileExists(t, path.Join(dir, ".tfdr/config.yaml"))
}
