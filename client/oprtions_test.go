package client

import (
	"testing"

	"github.com/cmd-stream/core-go"
)

func TestOptions(t *testing.T) {
	var (
		o                                     = Options{}
		wantCallback UnexpectedResultCallback = func(seq core.Seq, result core.Result) {}
	)
	Apply([]SetOption{WithUnexpectedResultCallback(wantCallback)}, &o)

	if o.UnexpectedResultCallback == nil {
		t.Errorf("UnexpectedResultCallback == nil")
	}
}
