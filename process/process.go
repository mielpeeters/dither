package process

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/mielpeeters/dither/colorpalette"
	"github.com/mielpeeters/dither/geom"
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
func Downscale(pixels *[][]color.Color, factor int) {
	fmt.Println(cyan + bold + "Downscaling..." + reset)
	iLen := len(*pixels)
	jLen := len((*pixels)[0])

	//create new image
	newImage := make([][]color.Color, roundDown(float64(iLen/factor)))
	for i := 0; i < len(newImage); i++ {
		newImage[i] = make([]color.Color, roundDown(float64(jLen/factor)))
	}

	wg := sync.WaitGroup{}

	for i := 0; i < iLen/factor; i++ {
		wg.Add(1) //do each row in parallel
		go func(i int) {
			for j := 0; j < jLen/factor; j++ {

				sumR := float64(0)
				sumG := float64(0)
				sumB := float64(0)
				sumA := float64(0)

				for k := 0; k < factor && i*factor+k <= iLen; k++ {
					for m := 0; m < factor && j*factor+m <= jLen; m++ {
						pixel := (*pixels)[i*factor+k][j*factor+m]

						originalColor, ok := color.RGBAModel.Convert(pixel).(color.RGBA)
						if !ok {
							fmt.Println("type conversion went wrong")
						}
						sumR += float64(originalColor.R)
						sumG += float64(originalColor.G)
						sumB += float64(originalColor.B)
						sumA += float64(originalColor.A)
					}
				}

				col := color.RGBA{
					uint8(sumR / math.Pow(float64(factor), 2)),
					uint8(sumG / math.Pow(float64(factor), 2)),
					uint8(sumB / math.Pow(float64(factor), 2)),
					uint8(sumA / math.Pow(float64(factor), 2)),
				}

				newImage[i][j] = col

			}
			wg.Done()
		}(i)

	}
	wg.Wait()
	*pixels = newImage
	fmt.Println(green + itallic + "	Done!" + reset)
}

// Upscale scales the input image up with the given integer factor
func Upscale(pixels *[][]color.Color, factor int) {
	fmt.Println(cyan + bold + "Upscaling..." + reset)
	ppixels := *pixels
	iLen := len(ppixels)
	jLen := len(ppixels[0])

	//create new image
	newImage := make([][]color.Color, roundDown(float64(iLen*factor)))
	for i := 0; i < len(newImage); i++ {
		newImage[i] = make([]color.Color, roundDown(float64(jLen*factor)))
	}

	wg := sync.WaitGroup{}

	for i := 0; i < iLen; i++ {
		wg.Add(1) //do each row in parallel
		go func(i int) {
			for j := 0; j < jLen; j++ {

				pixel := ppixels[i][j]

				for k := 0; k < factor; k++ {
					for l := 0; l < factor; l++ {
						newImage[i*factor+k][j*factor+l] = pixel
					}
				}

			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	*pixels = newImage
	fmt.Println(green + itallic + "	Done!" + reset)
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

func getColorDifference(left color.Color, right color.Color) errorColor {
	leftRGBA := colorpalette.ToRGBA(left)
	rightRGBA := colorpalette.ToRGBA(right)

	col := errorColor{
		R: int16(leftRGBA.R) - int16(rightRGBA.R),
		G: int16(leftRGBA.G) - int16(rightRGBA.G),
		B: int16(leftRGBA.B) - int16(rightRGBA.B),
		A: int16(leftRGBA.A) - int16(rightRGBA.A),
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

// FloydSteinbergDithering applies the FS dithering effect
//
// Returns a image.Paletted image pointer.
func FloydSteinbergDithering(pixels *[][]color.Color, palette colorpalette.ColorPalette) *image.Paletted {
	fmt.Println(cyan + bold + "Starting dithering process!" + reset)

	yLen := len(*pixels)
	xLen := len((*pixels)[0])

	upLeft := image.Point{0, 0}
	lowRight := image.Point{yLen, xLen}
	r := image.Rectangle{upLeft, lowRight}

	p := palette.ToPalette()

	newImage := image.NewPaletted(r, p)

	for y := 0; y < yLen; y++ {
		for x := 0; x < xLen; x++ {
			oldPixel := (*pixels)[y][x]

			colorIndex := uint8(p.Index(oldPixel))
			(*pixels)[y][x] = p[colorIndex]

			err := getColorDifference(oldPixel, (*pixels)[y][x])

			newImage.Set(y, x, oldPixel)

			if x+1 < xLen {
				(*pixels)[y][x+1] = addErrorToColor(err, (*pixels)[y][x+1], 7.0/16.0)
			}
			if x-1 >= 0 && y+1 < yLen {
				(*pixels)[y+1][x-1] = addErrorToColor(err, (*pixels)[y+1][x-1], 3.0/16.0)
			}
			if y+1 < yLen {
				(*pixels)[y+1][x] = addErrorToColor(err, (*pixels)[y+1][x], 5.0/16.0)
			}
			if x+1 < xLen && y+1 < yLen {
				(*pixels)[y+1][x+1] = addErrorToColor(err, (*pixels)[y+1][x+1], 1.0/16.0)
			}
		}
	}

	fmt.Println(green + itallic + "	Done!" + reset)

	return newImage
}

// ApplyErrorDiffusion will apply the error diffusion dithering, with the provided slice of
// error spreading ErrorDiffuser elements.
func ApplyErrorDiffusion(pixels *[][]color.Color, palette colorpalette.ColorPalette, diffusers *ErrorDiffusionMatrix) *image.Paletted {
	yLen := len(*pixels)
	xLen := len((*pixels)[0])

	upLeft := image.Point{0, 0}
	lowRight := image.Point{yLen, xLen}
	rect := image.Rectangle{upLeft, lowRight}

	pltte := palette.ToPalette()

	newImage := image.NewPaletted(rect, pltte)

	for y := 0; y < yLen; y++ {
		for x := 0; x < xLen; x++ {
			oldPixel := (*pixels)[y][x]

			colorIndex := uint8(pltte.Index(oldPixel))
			(*pixels)[y][x] = pltte[colorIndex]

			err := getColorDifference(oldPixel, (*pixels)[y][x])

			newImage.Set(y, x, oldPixel)

			for _, dif := range *diffusers {
				if dif.checkRange(x, y, xLen, yLen) {
					(*pixels)[y+dif.y][x+dif.x] = addErrorToColor(err, (*pixels)[y+dif.y][x+dif.x], dif.fraction)
				}
			}
		}
	}

	return newImage
}

func (dif *ErrorDiffuser) checkRange(x, y, xLen, yLen int) bool {

	if !((0 <= x+dif.x) && (x+dif.x < xLen)) {
		return false
	}
	if !((0 <= y+dif.y) && (y+dif.y < yLen)) {
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
