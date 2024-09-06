package proto_test

import (
	"os"
	"testing"

	"github.com/chengchung/nscard/proto"
)

func TestParsePlayHistory(t *testing.T) {
	bytes, err := os.ReadFile("./doc/example_playhistory.json")
	if err != nil {
		t.Fatal(err)
	}

	history, err := proto.ParsePlayHistory(bytes)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v", history)
}

func TestParseJPGameTitleInfoList(t *testing.T) {
	bytes, err := os.ReadFile("./doc/example_switch.xml")
	if err != nil {
		t.Fatal(err)
	}

	list, err := proto.ParseJPGameTitleInfoList(bytes)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v", list)
}
