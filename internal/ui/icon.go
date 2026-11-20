package ui

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"math"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"github.com/CriptoRochy/Karnize/internal/guard"
)

// Tone is the colour of the icon.
type Tone int

const (
	Gray Tone = iota
	Green
	Amber
	Red
)

var tones = map[Tone]color.NRGBA{
	Gray:  {0x9c, 0xa3, 0xaf, 0xff},
	Green: {0x16, 0xa3, 0x4a, 0xff},
	Amber: {0xea, 0x8a, 0x00, 0xff},
	Red:   {0xdc, 0x26, 0x26, 0xff},
}

func ToneOf(s guard.Snapshot) Tone {
	if s.Paused() {
		return Gray
	}
	switch s.Status {
	case guard.OK:
		return Green
	case guard.Suspect:
		return Amber
	case guard.Blocked:
		return Red
	}
	return Gray
}

// coverage of a pixel by an anti-aliased rounded rectangle (r = corner radius).
func roundRectAlpha(x, y, w, h, r float64) float64 {
	cx := math.Max(math.Max(r-x, x-(w-r)), 0)
	cy := math.Max(math.Max(r-y, y-(h-r)), 0)
	d := math.Hypot(cx, cy) - r
	return math.Min(math.Max(0.5-d, 0), 1)
}

func fill(img *image.NRGBA, c color.NRGBA, alpha func(x, y float64) float64) {
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			a := alpha(float64(x)+0.5, float64(y)+0.5)
			if a > 0 {
				cc := c
				cc.A = uint8(float64(c.A) * a)
				img.SetNRGBA(x, y, cc)
			}
		}
	}
}

// Dot is the macOS menu bar icon: a coloured circle, the text sits next to it.
func Dot(t Tone) []byte {
	const n = 32
	img := image.NewNRGBA(image.Rect(0, 0, n, n))
	fill(img, tones[t], func(x, y float64) float64 {
		d := math.Hypot(x-n/2, y-n/2) - 9
		return math.Min(math.Max(0.5-d, 0), 1)
	})
	var b bytes.Buffer
	_ = png.Encode(&b, img)
	return b.Bytes()
}

var (
	faceMu sync.Mutex
	faces  = map[float64]font.Face{}
)

func face(size float64) font.Face {
	faceMu.Lock()
	defer faceMu.Unlock()
	if f, ok := faces[size]; ok {
		return f
	}
	ft, _ := opentype.Parse(gobold.TTF)
	f, _ := opentype.NewFace(ft, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
	faces[size] = f
	return f
}

// Badge is the tray icon on Windows and Linux: a coloured tile with the country code,
// so the country is visible at a glance even without a menu bar title.
func Badge(t Tone, cc string, n int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, n, n))
	fn := float64(n)
	fill(img, tones[t], func(x, y float64) float64 { return roundRectAlpha(x, y, fn, fn, fn*0.22) })
	if len(cc) != 2 {
		// No country yet: a white dot in the middle.
		fill(img, color.NRGBA{255, 255, 255, 255}, func(x, y float64) float64 {
			d := math.Hypot(x-fn/2, y-fn/2) - fn*0.16
			return math.Min(math.Max(0.5-d, 0), 1)
		})
		return img
	}
	f := face(fn * 0.62)
	adv := font.MeasureString(f, cc)
	m := f.Metrics()
	x := (fixed.I(n) - adv) / 2
	y := (fixed.I(n) + m.Ascent - m.Descent) / 2
	d := &font.Drawer{Dst: img, Src: image.NewUniform(color.White), Face: f, Dot: fixed.Point26_6{X: x, Y: y}}
	d.DrawString(cc)
	return img
}

func pngBytes(img image.Image) []byte {
	var b bytes.Buffer
	_ = png.Encode(&b, img)
	return b.Bytes()
}

// ICO packs PNG images into a .ico file (supported since Windows Vista).
func ICO(imgs ...*image.NRGBA) []byte {
	var hdr, body bytes.Buffer
	binary.Write(&hdr, binary.LittleEndian, [3]uint16{0, 1, uint16(len(imgs))})
	offset := 6 + 16*len(imgs)
	for _, img := range imgs {
		p := pngBytes(img)
		w := img.Bounds().Dx()
		if w >= 256 {
			w = 0
		}
		binary.Write(&hdr, binary.LittleEndian, struct {
			W, H, Colors, Reserved uint8
			Planes, BPP            uint16
			Size, Offset           uint32
		}{uint8(w), uint8(w), 0, 0, 1, 32, uint32(len(p)), uint32(offset)})
		body.Write(p)
		offset += len(p)
	}
	return append(hdr.Bytes(), body.Bytes()...)
}

// AppIcon is the big icon for the .app bundle, the .exe and the README.
// Karnize's mark: a barrier - two posts and a bar - with a green light above it.
func AppIcon(n int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, n, n))
	fn := float64(n)
	bg := color.NRGBA{0x0b, 0x3b, 0x3f, 0xff} // deep teal
	m := fn * 0.09                            // macOS icon grid margin
	fill(img, bg, func(x, y float64) float64 { return roundRectAlpha(x-m, y-m, fn-2*m, fn-2*m, (fn-2*m)*0.225) })
	// rect draws an anti-aliased rounded bar given in icon units (0..1).
	rect := func(x0, y0, w, h float64) func(x, y float64) float64 {
		return func(x, y float64) float64 {
			return roundRectAlpha(x-x0*fn, y-y0*fn, w*fn, h*fn, math.Min(w, h)*fn/2)
		}
	}
	white := color.NRGBA{0xf1, 0xf5, 0xf9, 0xff}
	fill(img, white, rect(0.27, 0.46, 0.07, 0.30)) // left post
	fill(img, white, rect(0.66, 0.46, 0.07, 0.30)) // right post
	fill(img, white, rect(0.21, 0.50, 0.58, 0.08)) // the bar
	fill(img, tones[Green], func(x, y float64) float64 {
		d := math.Hypot(x-fn/2, y-fn*0.34) - fn*0.075
		return math.Min(math.Max(0.5-d, 0), 1)
	})
	return img
}
