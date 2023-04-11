// Package particled implements a new image type that stores each (original)
// pixel with constant colour at variable location.
// This enables colour-physics systems to be implemented on top.
package particled

import (
	"fmt"
	"image"
	"image/color"
)

// Pixicle is a Pixel Particle
// It has a constant colour, constant mass, and variable speed.
// The colour is stored as an index of some color.palette
type Pixicle struct {
	// Colour, contains the index of some palette
	Colour int
	// Velocity is the 2d speed that the particle currently has
	Velocity [2]float64
	// Mass is the constant movement inertia value of this Pixicle.
	Mass float64
	// Position stores the high-accuracy location of the Pixicle.
	Position    [2]float64
	newPosition [2]float64
	newVelocity [2]float64
	id          int
}

// simple 2d integer coordinate struct.
type coordinate struct {
	x, y int
}

// collection of pixicles at one location.
type colourClub struct {
	pixicles []color.Color
}

// Particled is a type of image that stores pixels at variable locations.
// Each pixel is represented as a Pixicle.
type Particled struct {
	// Particles is a pointer to a slice of all pixicles
	Pixicles *[]Pixicle
	// Palette holds the used colourpalette for this image
	Palette       color.Palette
	width, height int
}

func (pix Pixicle) toCoordinate() coordinate {
	return coordinate{
		x: int(pix.Position[0]),
		y: int(pix.Position[1]),
	}
}

func (pix Pixicle) equals(p Pixicle) bool {
	return pix.id == p.id
}

func addColour(one, two *color.RGBA, factor float64) {
	*one = color.RGBA{
		R: uint8(float64(one.R+two.R) / factor),
		G: uint8(float64(one.G+two.G) / factor),
		B: uint8(float64(one.B+two.B) / factor),
		A: uint8(float64(one.A+two.A) / factor),
	}
}

func (cc colourClub) average(palette *color.Palette) color.Color {
	// start with black
	var average color.RGBA = color.RGBA{
		R: 0,
		G: 0,
		B: 0,
		A: 0,
	}
	var add color.RGBA
	var ok bool

	// add colours iteratively
	for _, colour := range cc.pixicles {
		add, ok = color.RGBAModel.Convert(colour).(color.RGBA)
		if !ok {
			fmt.Println("Type conversion between colour models failed...")
		}
		addColour(&average, &add, 1.0/float64(len(cc.pixicles)))
	}

	return average
}

func (cc colourClub) add(value color.Color) {
	cc.pixicles = append(cc.pixicles, value)
}

// ColorModel returns the Image's color model.
func (p Particled) ColorModel() color.Model {
	return p.Palette
}

// Bounds returns the domain for which At can return non-zero color.
func (p Particled) Bounds() image.Rectangle {
	return image.Rectangle{
		Min: image.Point{
			X: 0,
			Y: 0,
		},
		Max: image.Point{
			X: p.width,
			Y: p.height,
		},
	}
}

// ToPaletted converts the Particled image back to a Paletted image.
func (p Particled) ToPaletted() *image.Paletted {
	paletted := image.NewPaletted(p.Bounds(), p.Palette)

	colourClubs := make(map[coordinate]colourClub)

	for _, pixicle := range *p.Pixicles {
		colourClubs[pixicle.toCoordinate()].add(p.Palette[pixicle.Colour])
	}

	for x := 0; x <= p.width; x++ {
		for y := 0; y <= p.height; y++ {
			paletted.Set(x, y, colourClubs[coordinate{x: x, y: y}].average(&p.Palette))
		}
	}

	return paletted
}

func (pix Pixicle) calculate(timestep float64, pixicles *[]Pixicle) {
	// TODO: define some physics rules and calculate the resulting position and velocity change
	// ---> use Runge Kutta 4th order for stability measures!
}

func (pix Pixicle) update() {
	// TODO: set all positions and velocities to the next ones.
}

// Iterate runs through one timestep of the physics loop.
func (p Particled) Iterate(timestep float64) {
	for _, pixicle := range *p.Pixicles {
		pixicle.calculate(timestep, p.Pixicles)
	}

	for _, pixicle := range *p.Pixicles {
		pixicle.update()
	}
}
