package x11

import (
	"testing"

	"github.com/jezek/xgb/xproto"
)

func TestParseModifiers(t *testing.T) {
	tests := []struct {
		name    string
		names   []string
		want    uint16
		wantErr bool
	}{
		{
			name:  "single",
			names: []string{"super"},
			want:  xproto.KeyButMaskMod4,
		},
		{
			name:  "multiple",
			names: []string{"super", "shift", "ctrl"},
			want:  xproto.KeyButMaskMod4 | xproto.KeyButMaskShift | xproto.KeyButMaskControl,
		},
		{
			name:  "case and whitespace",
			names: []string{"  SUPER ", "Shift"},
			want:  xproto.KeyButMaskMod4 | xproto.KeyButMaskShift,
		},
		{
			name:  "empty names ignored",
			names: []string{"", "  "},
			want:  0,
		},
		{
			name:    "unknown",
			names:   []string{"hyper"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseModifiers(tt.names...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseModifiers(%v) error = %v, wantErr %v", tt.names, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseModifiers(%v) = %#x, want %#x", tt.names, got, tt.want)
			}
		})
	}
}

func TestParseModifierString(t *testing.T) {
	tests := []struct {
		in      string
		want    uint16
		wantErr bool
	}{
		{in: "super+shift", want: xproto.KeyButMaskMod4 | xproto.KeyButMaskShift},
		{in: " ctrl + alt ", want: xproto.KeyButMaskControl | xproto.KeyButMaskMod1},
		{in: "", want: 0},
		{in: "   ", want: 0},
		{in: "super+bogus", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ParseModifierString(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseModifierString(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseModifierString(%q) = %#x, want %#x", tt.in, got, tt.want)
			}
		})
	}
}
