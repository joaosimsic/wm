package wm

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/BurntSushi/xgb/xproto"

	"wm/pkg/config"
)

func (wm *WM) allocColors() error {
	colormap := wm.screen.DefaultColormap

	type pair struct {
		ptr *uint32
		hex string
	}

	pairs := []pair{
		{&wm.barBg, wm.conf.Colors.BarBg},
		{&wm.barFg, wm.conf.Colors.BarFg},
		{&wm.barActiveBg, wm.conf.Colors.BarActiveBg},
		{&wm.borderActive, wm.conf.Colors.BorderActive},
		{&wm.borderInactive, wm.conf.Colors.BorderInactive},
	}

	for _, p := range pairs {
		if p.hex == "" {
			continue
		}
		val, err := config.HexToRGB(p.hex)
		if err != nil {
			continue
		}
		reply, err := xproto.AllocColor(wm.xu, colormap,
			uint16(val>>16), uint16((val>>8)&0xff), uint16(val&0xff)).Reply()
		if err != nil {
			return fmt.Errorf("alloc color %s: %w", p.hex, err)
		}
		*p.ptr = reply.Pixel
	}

	if wm.barBg == 0 {
		wm.barBg = wm.screen.BlackPixel
	}
	if wm.barFg == 0 {
		wm.barFg = wm.screen.WhitePixel
	}
	if wm.borderInactive == 0 {
		wm.borderInactive = wm.screen.BlackPixel
	}
	if wm.borderActive == 0 {
		wm.borderActive = wm.screen.WhitePixel
	}
	if wm.barActiveBg == 0 {
		wm.barActiveBg = wm.barBg
	}

	return nil
}

func (wm *WM) reloadConfig() error {
	if _, err := toml.DecodeFile("config.toml", &wm.conf); err != nil {
		return fmt.Errorf("reload config: %w", err)
	}
	return nil
}
