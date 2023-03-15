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

func getPalettesFromJSON(jsonFileName string) []ColorPalette {
	file, _ := ioutil.ReadFile(jsonFileName)

	data := []ColorPalette{}

	_ = json.Unmarshal([]byte(file), &data)

	return data
}

func getPaletteWithName(name string, palettes []ColorPalette) ColorPalette {
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

func paletteToJSONFile(palette ColorPalette, jsonFileName string) {
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

func paletteToImage(palette ColorPalette, fileName string) {
	pixels := make([][]color.Color, len(palette.Colors))

	for i := 0; i < len(palette.Colors); i++ {
		pixels[i] = make([]color.Color, 1)
	}

	for i := 0; i < len(palette.Colors); i++ {
		col := color.RGBA{
			uint8(palette.Colors[i][0]),
			uint8(palette.Colors[i][1]),
			uint8(palette.Colors[i][2]),
			uint8(palette.Colors[i][3]),
		}
		pixels[i][0] = col
	}

	upscale(&pixels, 10)
	image := pixelsToImage(&pixels)
	savePNG(image, fileName)
}

func convRGBAtoHSLA(rgba []float64) []float64 {
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

func convHSLAtoRGBA(hsla []float64) []float64 {
	h := hsla[0]
	s := hsla[1] / 100
	l := hsla[2] / 100
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
