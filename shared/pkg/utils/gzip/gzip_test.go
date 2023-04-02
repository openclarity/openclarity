package gzip

import (
	"os"
	"testing"

	"gotest.tools/assert"
)

func TestCompress(t *testing.T) {
	file, err := os.ReadFile("testdata/examples-bookinfo-reviews-v1:1.17.0.cdx.json")
	assert.NilError(t, err)
	t.Logf("file len %v", len(file))
	compress, err := Compress(file)
	assert.NilError(t, err)
	t.Logf("compress file len %v", len(compress))
	t.Logf("compress by %v", float64(len(compress))/float64(len(file))*100)
	fileB, err := UnCompress(compress)
	assert.NilError(t, err)
	t.Logf("fileB len %v", len(fileB))
	assert.DeepEqual(t, file, fileB)
}
