package main

import (
	"reflect"
	"testing"
)

func TestGetDurations(t *testing.T) {
	expected := []dir_size{
		{"../video_test/1", 30.447533},
		{"../video_test/2", 51.335066},
		{"../video_test/2/6", 20.887533},
		{"../video_test/3", 81.782599},
		{"../video_test/4", 72.222599},
		{"../video_test/5", 9.56},
		{"../video_test/8", 20.887533},
		{"Total in ../video_test/", 296.68286300000005},
	}
	dir := "../video_test/"
	summarize := false
	sort := false
	reverse := false
	vdu := NewVdu(dir)
	durations, _, _ := vdu.get_durations(summarize, sort, reverse)
	if !reflect.DeepEqual(durations, expected) {
		t.Errorf("get_durations(%s) != expected\nGot: %#v\nExpected: %#v", dir, durations, expected)
	}
}
