package next

import (
	"testing"
)

var js = []byte(`{
	"test": {
		"string_array": ["asdf", "ghjk", "zxcv"],
		"string_array_null": ["abc", null, "efg"],
		"array": [1, "2", 3],
		"arraywithsubs": [{"subkeyone": 1},
		{"subkeytwo": 2, "subkeythree": 3}],
		"int": 10,
		"float": 5.150,
		"string": "simplejson",
		"bool": true,
		"sub_obj": {"a": "1"}
	}
}`)

func TestRead(t *testing.T) {
	cfg := NewConfig()
	_, err := cfg.Read(js)

	if err != nil {
		t.Error("read json config fail")
	}
	t.Logf("sussess")
}

func TestGet(t *testing.T) {
	cfg := NewConfig()
	_, err := cfg.Read(js)

	if err != nil {
		t.Error("read json config fail")
	}

	str := cfg.String("test.string")
	if str != "simplejson" {
		t.Error("get config node fail")
	}
	t.Log("JSON value: test.string -> ", str)
}

func TestGet2(t *testing.T) {
	cfg := NewConfig()
	_, err := cfg.Read(js)

	if err != nil {
		t.Error("read json config fail")
	}

	str := cfg.String("test.sub_obj.a")
	if str != "1" {
		t.Error("get test.sub_obj.a fail")
	}
	t.Log("JSON value: test.sub_obj.a -> ", str)
}

func TestSet(t *testing.T) {
	cfg := NewConfig()
	_, err := cfg.Read(js)

	if err != nil {
		t.Error("read json config fail")
	}

	cfg.Set("test.sub_obj.b", "2")
	str := cfg.String("test.sub_obj.b")
	if str != "2" {
		t.Error("get config node fail")
	}
	t.Log("JSON value: test.sub_obj.b -> ", str)
}
