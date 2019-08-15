package localconfig

import (
	"testing"
	"github.com/davecgh/go-spew/spew"
)

func TestLoad(t *testing.T) {
	conf, _ := Load()
	spew.Dump(conf)
	t.Logf("%s\n",conf.Peer.Abcd)
}
