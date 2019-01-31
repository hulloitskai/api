package common

import (
	ms "github.com/mitchellh/mapstructure"
)

var (
	// DecoderConfigOption is the common viper.DecoderConfigOption used by
	// Tesseract packages.
	DecoderConfigOption = func(dc *ms.DecoderConfig) {
		dc.TagName = "ms"
	}
)
