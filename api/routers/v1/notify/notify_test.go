package notify

import (
	"strconv"
	"testing"
)

func TestParseUIDs(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want []uint32
		ok   bool
	}{
		{"single", "5", []uint32{5}, true},
		{"multiple", "1,2,3", []uint32{1, 2, 3}, true},
		{"spaces and dedupe", " 1 , 2 ,1 ", []uint32{1, 2}, true},
		{"empty", "", nil, false},
		{"only commas", ",, ,", nil, false},
		{"zero id", "0", nil, false},
		{"non numeric", "1,abc", nil, false},
		{"negative", "-1", nil, false},
		{"overflow uint32", "4294967296", nil, false},
	}
	for _, c := range cases {
		uids, ok := parseUIDs(c.raw)
		if ok != c.ok {
			t.Errorf("%s: parseUIDs(%q) ok = %v, want %v", c.name, c.raw, ok, c.ok)
			continue
		}
		if len(uids) != len(c.want) {
			t.Errorf("%s: parseUIDs(%q) = %v, want %v", c.name, c.raw, uids, c.want)
			continue
		}
		for i := range uids {
			if uids[i] != c.want[i] {
				t.Errorf("%s: parseUIDs(%q) = %v, want %v", c.name, c.raw, uids, c.want)
				break
			}
		}
	}

	// 超过单次查询上限（去重后计数）
	raw := ""
	for i := 1; i <= 101; i++ {
		if i > 1 {
			raw += ","
		}
		raw += strconv.Itoa(i)
	}
	if _, ok := parseUIDs(raw); ok {
		t.Error("parseUIDs should reject more than 100 uids")
	}

	// 去重后不超上限则放行
	if uids, ok := parseUIDs("1,1,1"); !ok || len(uids) != 1 {
		t.Errorf("parseUIDs dedupe = %v, %v; want [1], true", uids, ok)
	}
}
