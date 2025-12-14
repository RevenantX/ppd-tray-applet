package assets

import _ "embed"

//go:embed perf.png
var performance []byte

//go:embed balanced.png
var balanced []byte

//go:embed powersave.png
var powersave []byte

var Images = map[string][]byte{
	"performance": performance,
	"balanced":    balanced,
	"powersave":   powersave,
	"power-saver": powersave,
	"power-save":  powersave,
}
