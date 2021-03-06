package canvas

import (
	"image"

	"golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
	"golang.org/x/image/vector"
)

type Rasterizer struct {
	img draw.Image
	dpm float64
}

// NewRasterizer creates a renderer that draws to a rasterized image.
func NewRasterizer(img draw.Image, dpm float64) *Rasterizer {
	return &Rasterizer{
		img: img,
		dpm: dpm,
	}
}

func (r *Rasterizer) Size() (float64, float64) {
	size := r.img.Bounds().Size()
	return float64(size.X) / r.dpm, float64(size.Y) / r.dpm
}

func (r *Rasterizer) RenderPath(path *Path, style Style, m Matrix) {
	// TODO: use fill rule (EvenOdd, NonZero) for rasterizer
	path = path.Transform(m)

	strokeWidth := 0.0
	if style.StrokeColor.A != 0 && 0.0 < style.StrokeWidth {
		strokeWidth = style.StrokeWidth
	}

	size := r.img.Bounds().Size()
	bounds := path.Bounds()
	dx, dy := 0, 0
	x := int((bounds.X - strokeWidth) * r.dpm)
	y := int((bounds.Y - strokeWidth) * r.dpm)
	w := int((bounds.W+2*strokeWidth)*r.dpm) + 1
	h := int((bounds.H+2*strokeWidth)*r.dpm) + 1
	if (x+w <= 0 || size.X <= x) && (y+h <= 0 || size.Y <= y) {
		return // outside canvas
	}

	if x < 0 {
		dx = -x
		x = 0
	}
	if y < 0 {
		dy = -y
		y = 0
	}
	if size.X <= x+w {
		w = size.X - x
	}
	if size.Y <= y+h {
		h = size.Y - y
	}
	if w <= 0 || h <= 0 {
		return // has no size
	}

	path = path.Translate(-float64(x)/r.dpm, -float64(y)/r.dpm)
	if style.FillColor.A != 0 {
		ras := vector.NewRasterizer(w, h)
		path.ToRasterizer(ras, r.dpm)
		ras.Draw(r.img, image.Rect(x, size.Y-y, x+w, size.Y-y-h), image.NewUniform(style.FillColor), image.Point{dx, dy})
	}
	if style.StrokeColor.A != 0 && 0.0 < style.StrokeWidth {
		if 0 < len(style.Dashes) {
			path = path.Dash(style.DashOffset, style.Dashes...)
		}
		path = path.Stroke(style.StrokeWidth, style.StrokeCapper, style.StrokeJoiner)

		ras := vector.NewRasterizer(w, h)
		path.ToRasterizer(ras, r.dpm)
		ras.Draw(r.img, image.Rect(x, size.Y-y, x+w, size.Y-y-h), image.NewUniform(style.StrokeColor), image.Point{dx, dy})
	}
}

func (r *Rasterizer) RenderText(text *Text, m Matrix) {
	paths, colors := text.ToPaths()
	for i, path := range paths {
		style := DefaultStyle
		style.FillColor = colors[i]
		r.RenderPath(path, style, m)
	}
}

func (r *Rasterizer) RenderImage(img image.Image, m Matrix) {
	origin := m.Dot(Point{0, float64(img.Bounds().Size().Y)}).Mul(r.dpm)
	m = m.Scale(r.dpm, r.dpm)

	h := float64(r.img.Bounds().Size().Y)
	aff3 := f64.Aff3{m[0][0], -m[0][1], origin.X, -m[1][0], m[1][1], h - origin.Y}

	// add transparent margin to image for smooth borders when rotating
	margin := 4
	size := img.Bounds().Size()
	img2 := image.NewRGBA(image.Rect(0, 0, size.X+margin*2, size.Y+margin*2))
	draw.Draw(img2, image.Rect(margin, margin, size.X, size.Y), img, image.Point{}, draw.Over)

	draw.CatmullRom.Transform(r.img, aff3, img2, img2.Bounds(), draw.Over, nil)
}
