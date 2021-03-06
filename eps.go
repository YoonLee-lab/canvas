package canvas

import (
	"fmt"
	"image"
	"image/color"
	"io"
)

var psEllipseDef = `/ellipse {
/rot exch def
/endangle exch def
/startangle exch def
/yrad exch def
/xrad exch def
/y exch def
/x exch def
/savematrix matrix currentmatrix def
x y translate
rot rotate
xrad yrad scale
0 0 1 startangle endangle arc
savematrix setmatrix
} def /ellipsen {
/rot exch def
/endangle exch def
/startangle exch def
/yrad exch def
/xrad exch def
/y exch def
/x exch def
/savematrix matrix currentmatrix def
x y translate
rot rotate
xrad yrad scale
0 0 1 startangle endangle arcn
savematrix setmatrix
} def`

type EPS struct {
	w             io.Writer
	width, height float64
	color         color.RGBA
}

// NewEPS creates an encapsulated PostScript renderer.
func NewEPS(w io.Writer, width, height float64) *EPS {
	fmt.Fprintf(w, "%%!PS-Adobe-3.0 EPSF-3.0\n%%%%BoundingBox: 0 0 %v %v\n", dec(width), dec(height))
	fmt.Fprintf(w, psEllipseDef)
	// TODO: (EPS) generate and add preview

	return &EPS{
		w:      w,
		width:  width,
		height: height,
		color:  Black,
	}
}

func (r *EPS) setColor(color color.RGBA) {
	if color != r.color {
		fmt.Fprintf(r.w, " %v %v %v setrgbcolor", dec(float64(color.R)/255.0), dec(float64(color.G)/255.0), dec(float64(color.B)/255.0))
		r.color = color
	}
}

func (r *EPS) Size() (float64, float64) {
	return r.width, r.height
}

func (r *EPS) RenderPath(path *Path, style Style, m Matrix) {
	// TODO: (EPS) test ellipse, rotations etc
	// TODO: (EPS) add drawState support
	// TODO: (EPS) use dither to fake transparency
	r.setColor(style.FillColor)
	r.w.Write([]byte(" "))
	r.w.Write([]byte(path.Transform(m).ToPS()))
	r.w.Write([]byte(" fill"))
}

func (r *EPS) RenderText(text *Text, m Matrix) {
	// TODO: (EPS) write text natively
	paths, colors := text.ToPaths()
	for i, path := range paths {
		style := DefaultStyle
		style.FillColor = colors[i]
		r.RenderPath(path, style, m)
	}
}

func (r *EPS) RenderImage(img image.Image, m Matrix) {
	// TODO: (EPS) write image
}
