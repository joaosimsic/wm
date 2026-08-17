package x11

import (
	"reflect"
	"testing"

	"github.com/jezek/xgb/xproto"
)

func TestCombos(t *testing.T) {
	tests := []struct {
		name    string
		mods    uint16
		numlock uint16
		want    []uint16
	}{
		{
			name:    "no overlap",
			mods:    xproto.KeyButMaskMod4,
			numlock: xproto.KeyButMaskMod2,
			want: []uint16{
				xproto.KeyButMaskMod4,
				xproto.KeyButMaskMod4 | xproto.KeyButMaskLock,
				xproto.KeyButMaskMod4 | xproto.KeyButMaskMod2,
				xproto.KeyButMaskMod4 | xproto.KeyButMaskLock | xproto.KeyButMaskMod2,
			},
		},
		{
			name:    "numlock equals lock",
			mods:    xproto.KeyButMaskMod4,
			numlock: xproto.KeyButMaskLock,
			want: []uint16{
				xproto.KeyButMaskMod4,
				xproto.KeyButMaskMod4 | xproto.KeyButMaskLock,
			},
		},
		{
			name:    "no numlock",
			mods:    xproto.KeyButMaskMod4,
			numlock: 0,
			want: []uint16{
				xproto.KeyButMaskMod4,
				xproto.KeyButMaskMod4 | xproto.KeyButMaskLock,
			},
		},
		{
			name:    "mods already contains lock",
			mods:    xproto.KeyButMaskMod4 | xproto.KeyButMaskLock,
			numlock: xproto.KeyButMaskMod2,
			want: []uint16{
				xproto.KeyButMaskMod4 | xproto.KeyButMaskLock,
				xproto.KeyButMaskMod4 | xproto.KeyButMaskLock | xproto.KeyButMaskMod2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := combos(tt.mods, tt.numlock); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("combos(%#x, %#x) = %#v, want %#v", tt.mods, tt.numlock, got, tt.want)
			}
		})
	}
}
