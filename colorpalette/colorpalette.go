package colorpalette

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io/ioutil"
	"math"
)

// ColorPalette contains name and colors of one colorpalette
type ColorPalette struct {
	Name   string  `json:"name"`
	Colors [][]int `json:"colors"`
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
func GetPaletteWithName(name string, palettes []ColorPalette) ColorPalette {
	for _, pltt := range palettes {
		if pltt.Name == name {
			return pltt
		}
	}

	black := []int{0, 0, 0, 255}
	colors := [][]int{black}

	val := ColorPalette{
		"New",
		colors,
	}
	return val
}

// ToJSONFile writes the given ColorPalette out to the specified path, as a JSON file (formatted).
func (palette *ColorPalette) ToJSONFile(jsonFileName string) {
	output, err := json.MarshalIndent(palette, "", "  ")
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
func (palette *ColorPalette) ToPalette() color.Palette {
	colors := []color.Color{}
	var paletteColor color.Color

	for i := 0; i < len(palette.Colors); i++ {
		paletteColor = color.RGBA{
			uint8(palette.Colors[i][0]),
			uint8(palette.Colors[i][1]),
			uint8(palette.Colors[i][2]),
			uint8(palette.Colors[i][3]),
		}
		colors = append(colors, paletteColor)
	}

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
