package process

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/mielpeeters/dither/geom"
	"golang.org/x/image/draw"
)

type errorColor struct {
	R int16
	G int16
	B int16
	A int16
}

// ErrorDiffuser represents one spreaded error, with parameters
// x_offset, y_offset, and the fraction of the error to divide
type ErrorDiffuser struct {
	x        int
	y        int
	fraction float64
}

// AdjustableImage is an interface to define images that implement the .Set() function
type AdjustableImage interface {
	ColorModel() color.Model
	Bounds() image.Rectangle
	RGBAAt(x, y int) color.RGBA
	Set(x, y int, c color.Color)
}

// ErrorDiffusionMatrix is the matrix that is used to spread the errors
type ErrorDiffusionMatrix []ErrorDiffuser

// FloydSteinBerg is the EDM used for FS dithering
var FloydSteinBerg = *makeFloydSteinBerg()

// Simple is the EDM used for simple 2d dithering
var Simple = *makeSimpleDiffuser()

// Stucki is the EDM used for Stucki dithering
var Stucki = *makeStuckiDiffuser()

// JarvisJudiceNinke is the EDM used for JarvisJudiceNinke dithering
var JarvisJudiceNinke = ErrorDiffusionMatrix{
	{1, 0, 7.0 / 48.0},
	{2, 0, 5.0 / 48.0},
	{-2, 1, 3.0 / 48.0},
	{-1, 1, 5.0 / 48.0},
	{0, 1, 7.0 / 48.0},
	{1, 1, 5.0 / 48.0},
	{2, 1, 3.0 / 48.0},
	{-2, 2, 1.0 / 48.0},
	{-1, 2, 3.0 / 48.0},
	{0, 2, 5.0 / 48.0},
	{1, 2, 3.0 / 48.0},
	{2, 2, 1.0 / 48.0},
}

func roundDown(number float64) int {
	return int(math.Floor(number))
}

const reset = "\033[0m"
const cyan = "\033[36m"
const green = "\033[32m"
const itallic = "\033[3m"
const bold = "\033[1m"
const red = "\033[31m"
const blink = "\033[5m"

// Downscale scales the image down with a given integer factor
func Downscale(img image.Image, factor int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, img.Bounds().Max.X/factor, img.Bounds().Max.Y/factor))
	draw.NearestNeighbor.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)

	return dst
}

// Upscale scales the input image up with the given integer factor
func Upscale(img image.Image, factor int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, img.Bounds().Max.X*factor, img.Bounds().Max.Y*factor))
	draw.NearestNeighbor.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)

	return dst
}

func addColorComponents(left int16, right int16) uint8 {
	result := left + right

	if result < 0 {
		result = 0
	}

	if result > 255 {
		result = 255
	}

	return uint8(result)
}

func addErrorToColor(errorColor errorColor, origColor color.Color, factor float64) color.Color {
	orig, ok := color.RGBAModel.Convert(origColor).(color.RGBA)
	if !ok {
		fmt.Println("type conversion (to rgba color) went wrong")
	}

	col := color.RGBA{
		addColorComponents(int16(orig.R), int16(float64(errorColor.R)*factor)),
		addColorComponents(int16(orig.G), int16(float64(errorColor.G)*factor)),
		addColorComponents(int16(orig.B), int16(float64(errorColor.B)*factor)),
		addColorComponents(int16(orig.A), int16(float64(errorColor.A)*factor)),
	}

	// printRGBAColor(orig, "original neighboring color")
	// printRGBAColor(col, "added noise neighboring color")

	return col
}

func getColorDifference(left color.RGBA, right color.RGBA) errorColor {

	col := errorColor{
		R: int16(left.R) - int16(right.R),
		G: int16(left.G) - int16(right.G),
		B: int16(left.B) - int16(right.B),
		A: int16(left.A) - int16(right.A),
	}

	return col
}

func makeColor(R, G, B, A int) color.Color {
	col := color.RGBA{
		uint8(R),
		uint8(G),
		uint8(B),
		uint8(A),
	}

	return col
}

func pointToColor(point geom.Point) color.Color {
	//rgba := HSLAtoRGBA(point.Coordinates)
	col := color.RGBA{
		uint8(point.Coordinates[0]),
		uint8(point.Coordinates[1]),
		uint8(point.Coordinates[2]),
		uint8(point.Coordinates[3]),
	}

	return col
}

// ApplyErrorDiffusion will apply the error diffusion dithering, with the provided slice of
// error spreading ErrorDiffuser elements.
func ApplyErrorDiffusion(img AdjustableImage, palette color.Palette, diffusers *ErrorDiffusionMatrix) *image.Paletted {
	X := img.Bounds().Max.X
	Y := img.Bounds().Max.Y

	rect := img.Bounds()

	newImage := image.NewPaletted(rect, palette)

	for y := 0; y <= Y; y++ {
		for x := 0; x <= X; x++ {
			oldPixel := img.RGBAAt(x, y)

			colorIndex := uint8(palette.Index(oldPixel))

			img.Set(x, y, palette[colorIndex])

			err := getColorDifference(oldPixel, img.RGBAAt(x, y))

			// automatically assigns that index that corresponds with oldPixel the best!
			newImage.Set(x, y, oldPixel)

			for _, dif := range *diffusers {
				if dif.checkRange(x, y, X, Y) {
					img.Set(x+dif.x, y+dif.y, addErrorToColor(err, img.RGBAAt(x+dif.x, y+dif.y), dif.fraction))
				}
			}
		}
	}

	return newImage
}

func (dif *ErrorDiffuser) checkRange(x, y, X, Y int) bool {

	if !((0 <= x+dif.x) && (x+dif.x <= X)) {
		return false
	}
	if !((0 <= y+dif.y) && (y+dif.y <= Y)) {
		return false
	}
	return true
}

// makeFloydSteinBerg returns the correct error diffusion matrix struct for usage in ApplyErrorDiffusion
func makeFloydSteinBerg() *ErrorDiffusionMatrix {
	diffusers := make(ErrorDiffusionMatrix, 0)

	diffusers = append(diffusers, ErrorDiffuser{
		x:        1,
		y:        0,
		fraction: 7.0 / 16.0,
	})
	diffusers = append(diffusers, ErrorDiffuser{
		x:        -1,
		y:        1,
		fraction: 3.0 / 16.0,
	})
	diffusers = append(diffusers, ErrorDiffuser{
		x:        0,
		y:        1,
		fraction: 5.0 / 16.0,
	})
	diffusers = append(diffusers, ErrorDiffuser{
		x:        1,
		y:        1,
		fraction: 1.0 / 16.0,
	})

	return &diffusers
}

// makeSimpleDiffuser returns the simplest diffusers possible
func makeSimpleDiffuser() *ErrorDiffusionMatrix {
	diffusers := make(ErrorDiffusionMatrix, 0)

	diffusers = append(diffusers, ErrorDiffuser{
		x:        1,
		y:        0,
		fraction: 1.0 / 2.0,
	})
	diffusers = append(diffusers, ErrorDiffuser{
		x:        0,
		y:        1,
		fraction: 1.0 / 2.0,
	})

	return &diffusers
}

// makeStuckiDiffuser returns the simplest diffusers possible
func makeStuckiDiffuser() *ErrorDiffusionMatrix {
	diffusers := make(ErrorDiffusionMatrix, 0)

	diffusers = append(diffusers, ErrorDiffuser{
		x:        1,
		y:        0,
		fraction: 8.0 / 42.0,
	})
	diffusers = append(diffusers, ErrorDiffuser{
		x:        2,
		y:        0,
		fraction: 4.0 / 42.0,
	})
	diffusers = append(diffusers, ErrorDiffuser{
		x:        -2,
		y:        1,
		fraction: 2.0 / 42.0,
	})
	diffusers = append(diffusers, ErrorDiffuser{
		x:        -1,
		y:        4,
		fraction: 1.0 / 42.0,
	})
	diffusers = append(diffusers, ErrorDiffuser{
		x:        0,
		y:        1,
		fraction: 8.0 / 42.0,
	})
	diffusers = append(diffusers, ErrorDiffuser{
		x:        1,
		y:        1,
		fraction: 4.0 / 42.0,
	})
	diffusers = append(diffusers, ErrorDiffuser{
		x:        2,
		y:        1,
		fraction: 2.0 / 42.0,
	})
	diffusers = append(diffusers, ErrorDiffuser{
		x:        -2,
		y:        2,
		fraction: 1.0 / 42.0,
	})
	diffusers = append(diffusers, ErrorDiffuser{
		x:        -1,
		y:        2,
		fraction: 2.0 / 42.0,
	})
	diffusers = append(diffusers, ErrorDiffuser{
		x:        0,
		y:        2,
		fraction: 4.0 / 42.0,
	})
	diffusers = append(diffusers, ErrorDiffuser{
		x:        1,
		y:        2,
		fraction: 2.0 / 42.0,
	})
	diffusers = append(diffusers, ErrorDiffuser{
		x:        2,
		y:        2,
		fraction: 1.0 / 42.0,
	})

	return &diffusers
}
