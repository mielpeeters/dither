package colorpalette

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"math"

	"github.com/mielpeeters/dither/geom"
	"github.com/mielpeeters/dither/kmeans"
)

// ColorPalette contains name and colors of one colorpalette
type ColorPalette struct {
	Name   string  `json:"name"`
	Colors [][]int `json:"colors"`
}

// KMAccuracy is he accuracy needed for convergence of the k-means algorithm used in function Create
var KMAccuracy = 0.01

// KMConsecutive is the consecutive accuracy hits needed for convergence of the k-means algorithm
// used in function Create
var KMConsecutive = 2

// SampleFactor describes the fraction of pixels to be used in creating a palette.
var SampleFactor = 5

// KMTimes descibes how many times the Kmeans algorithm needs to be run with random start
var KMTimes = 3

// Create creates a new colorpalette using the k-means clustering algorithm
//
//   - samplefactor: how many pixles to skip, during sampling for the creatrion of the KMeans problem's cluster points
//     (higher means faster, because less points to iterate over)
//   - kmTimes defines the amount of times to start the k-means algorithm with random init, the best output is choosen
func Create(img image.Image, k int) color.Palette {
	pointSet := geom.PointSet{}
	// sample only 1/samplefactor of the pixels
	for x := 0; x < img.Bounds().Max.X; x += SampleFactor {
		for y := 0; y < img.Bounds().Max.Y; y += SampleFactor {
			newPoint := colorToPoint(img.At(x, y))
			newPoint.ID = x + y*img.Bounds().Max.X

			pointSet.Points = append(pointSet.Points, newPoint)
		}
	}

	var colorPalettes []ColorPalette
	var errors []float64

	// do the algorithm kmTimes
	for i := 0; i < KMTimes; i++ {
		KM := kmeans.CreateKMeansProblem(pointSet, k, geom.RedMeanDistance)

		KM.Cluster(KMAccuracy, KMConsecutive)

		colorPalette := ColorPalette{}
		for index := range KM.KMeans.Points {
			colorPalette.Colors = append(colorPalette.Colors, pointToColorSlice(KM.KMeans.Points[index]))
		}

		colorPalettes = append(colorPalettes, colorPalette)
		errors = append(errors, KM.TotalDist())
	}

	// now select the colorpalette with the lowest error!
	minIndex := findMinIndex(errors)

	return colorPalettes[minIndex].ToPalette()
}

// CreatePLT creates a new colorpalette using the k-means clustering algorithm
//
//   - samplefactor: how many pixles to skip, during sampling for the creatrion of the KMeans problem's cluster points
//     (higher means faster, because less points to iterate over)
//   - kmTimes defines the amount of times to start the k-means algorithm with random init, the best output is choosen
func CreatePLT(img image.Image, k int) ColorPalette {
	pointSet := geom.PointSet{}
	// sample only 1/samplefactor of the pixels
	for x := 0; x < img.Bounds().Max.X; x += SampleFactor {
		for y := 0; y < img.Bounds().Max.Y; y += SampleFactor {
			newPoint := colorToPoint(img.At(x, y))
			newPoint.ID = x + y*img.Bounds().Max.X

			pointSet.Points = append(pointSet.Points, newPoint)
		}
	}

	var colorPalettes []ColorPalette
	var errors []float64

	// do the algorithm kmTimes
	for i := 0; i < KMTimes; i++ {
		KM := kmeans.CreateKMeansProblem(pointSet, k, geom.RedMeanDistance)

		KM.Cluster(KMAccuracy, KMConsecutive)

		colorPalette := ColorPalette{}
		for index := range KM.KMeans.Points {
			colorPalette.Colors = append(colorPalette.Colors, pointToColorSlice(KM.KMeans.Points[index]))
		}

		colorPalettes = append(colorPalettes, colorPalette)
		errors = append(errors, KM.TotalDist())
	}

	// now select the colorpalette with the lowest error!
	minIndex := findMinIndex(errors)

	return colorPalettes[minIndex]
}

// Traverse is used to find colours on one line in the image
func (colorpalette *ColorPalette) Traverse(img *image.Image, ltr bool, name string) (ColorPalette, int) {
	var step image.Point
	var maxX, maxY int
	var currentPix image.Point
	palette := ColorPalette{
		Name:   name,
		Colors: [][]int{},
	}

	maxBounds := (*img).Bounds().Max
	maxX = maxBounds.X
	maxY = maxBounds.Y

	if ltr {
		step = image.Point{1, 0}
		currentPix = image.Point{
			X: 0,
			Y: maxBounds.Y / 2,
		}
	} else {
		step = image.Point{0, 1}
		currentPix = image.Point{
			X: maxBounds.X / 2,
			Y: 0,
		}
	}

	consecutiveSame := 0

	white := color.RGBA{
		R: 255,
		G: 255,
		B: 255,
		A: 255,
	}

	currentColor := color.RGBA{
		R: 255,
		G: 255,
		B: 255,
		A: 255,
	}

	for currentPix.X < maxX && currentPix.Y < maxY {
		clr := (*img).At(currentPix.X, currentPix.Y)
		clrRGBA := ToRGBA(clr)

		if clrRGBA == currentColor {
			consecutiveSame++
		} else {
			consecutiveSame = 0
			currentColor = clrRGBA
		}

		if consecutiveSame == 8 && currentColor != white {
			clrArray := []int{
				int(currentColor.R),
				int(currentColor.G),
				int(currentColor.B),
				int(currentColor.A),
			}
			palette.Colors = append(palette.Colors, clrArray)
		}

		currentPix = currentPix.Add(step)
	}

	return palette, len(palette.Colors)
}

func pointToColorSlice(point geom.Point) []int {
	returnValue := []int{}

	for _, value := range point.Coordinates {
		returnValue = append(returnValue, int(value))
	}

	return returnValue
}

func colorToPoint(clr color.Color) geom.Point {
	clrRGBA := ToRGBA(clr)
	coordinates := []float32{float32(clrRGBA.R), float32(clrRGBA.G), float32(clrRGBA.B), float32(clrRGBA.A)}
	//coordinates = RGBAtoHSLA(coordinates)
	point := geom.Point{
		Coordinates: coordinates,
		ID:          0,
	}
	return point
}

func findMinIndex(arr []float64) int {
	min := math.Inf(1) // Initialize min with the highest possible float64 value
	minIndex := 0      // Initialize minIndex with 0
	for i, v := range arr {
		if v < min {
			min = v
			minIndex = i
		}
	}
	return minIndex
}

// GetPalettesFromJSON returns a slice of ColorPalettes after reading them from a JSON file.
func GetPalettesFromJSON(jsonFileName string) []ColorPalette {
	file, _ := ioutil.ReadFile(jsonFileName)

	data := []ColorPalette{}

	_ = json.Unmarshal([]byte(file), &data)

	return data
}

// GetPaletteWithName returns a specific from a slice of ColorPalette.
// The palette is specified by name.
// If there is none that matches, a black ColorPalette is returned.
func GetPaletteWithName(name string, palettes []ColorPalette) *ColorPalette {
	for _, pltt := range palettes {
		if pltt.Name == name {
			return &pltt
		}
	}

	black := []int{0, 0, 0, 255}
	colors := [][]int{black}

	val := ColorPalette{
		"New",
		colors,
	}
	return &val
}

// ToJSONFile writes the given ColorPalette out to the specified path, as a JSON file (formatted).
func (colorpalette *ColorPalette) ToJSONFile(jsonFileName string) {
	output, err := json.MarshalIndent(colorpalette, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = ioutil.WriteFile(jsonFileName, output, 0644)

	if err != nil {
		fmt.Println(err)
	}
}

// ToJSONFileNoIndent writes the given ColorPalette out to the specified path, as a JSON file (formatted).
func (colorpalette *ColorPalette) ToJSONFileNoIndent(jsonFileName string) {
	output, err := json.Marshal(colorpalette)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = ioutil.WriteFile(jsonFileName, output, 0644)

	if err != nil {
		fmt.Println(err)
	}
}

// ConvRGBAtoHSLA converts between RGBA and HSLA color formats
func ConvRGBAtoHSLA(rgba []float64) []float64 {
	r := float64(rgba[0]) / 255.0
	g := float64(rgba[1]) / 255.0
	b := float64(rgba[2]) / 255.0
	a := float64(rgba[3])

	cMax := math.Max(math.Max(r, g), b)
	cMin := math.Min(math.Min(r, g), b)

	delta := cMax - cMin

	var hue float64
	var saturation float64
	var lightness float64

	if delta == 0 {
		hue = 0.0
		saturation = 0.0
	} else if cMax == r {
		hue = (g - b) / delta

		for hue > 6 { //apply modulo 6
			hue -= 6
		}

		hue = hue * 60
	} else if cMax == g {
		hue = 60 * ((b-r)/delta + 2)
	} else {
		hue = 60 * ((r-g)/delta + 4)
	}

	lightness = (cMax + cMin) / 2.0

	if delta != 0 {
		saturation = delta / (1 - math.Abs(2*lightness-1))
	}

	output := []float64{hue, saturation * 100, lightness * 100, a}

	return output
}

// ConvHSLAtoRGBA converts between HSLA and RGBA color formats
func ConvHSLAtoRGBA(hsla []float64) []float64 {
	h := hsla[0]
	s := hsla[1] / 100.0
	l := hsla[2] / 100.0
	a := hsla[3]

	c := (1 - math.Abs(2*l-1)) * s

	modVal := h / 60
	for modVal > 2 {
		modVal -= 2
	}
	x := c * (1 - math.Abs(modVal-1))
	m := l - c/2

	var r float64
	var g float64
	var b float64
	if 0 <= h && h < 60 {
		r = c
		g = x
	} else if 60 <= h && h < 120 {
		r = x
		g = c
	} else if 120 <= h && h < 180 {
		g = c
		b = x
	} else if 180 <= h && h < 240 {
		g = x
		b = c
	} else if 240 <= h && h < 300 {
		r = x
		b = c
	} else {
		r = c
		b = x
	}

	r = (r + m) * 255
	g = (g + m) * 255
	b = (b + m) * 255

	output := []float64{r, g, b, a}

	return output
}

// ToPalette converts between this custom ColorPalette and the
// Go standard library color.Palette type struct
func (colorpalette *ColorPalette) ToPalette() color.Palette {
	colors := []color.Color{}
	var paletteColor color.Color

	for i := 0; i < len(colorpalette.Colors); i++ {
		paletteColor = color.RGBA{
			uint8(colorpalette.Colors[i][0]),
			uint8(colorpalette.Colors[i][1]),
			uint8(colorpalette.Colors[i][2]),
			uint8(colorpalette.Colors[i][3]),
		}
		colors = append(colors, paletteColor)
	}

	return colors
}

// BW returns a black and white color palette
func BW() color.Palette {
	colors := []color.Color{}

	colors = append(colors, color.RGBA{
		R: 0,
		G: 0,
		B: 0,
		A: 255,
	})

	colors = append(colors, color.RGBA{
		R: 255,
		G: 255,
		B: 255,
		A: 255,
	})

	return colors
}

// ToRGBA converts some color.Color into color.RGBA
func ToRGBA(origColor color.Color) color.RGBA {
	orig, ok := color.RGBAModel.Convert(origColor).(color.RGBA)
	if !ok {
		fmt.Println("type conversion (to rgba color) went wrong")
	}
	return orig
}
