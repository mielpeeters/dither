// Package particled implements a new image type that stores each (original)
// pixel with constant colour at variable location.
// This enables colour-physics systems to be implemented on top.
package particled

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/mielpeeters/dither/geom"
)

// Pixicle is a Pixel Particle
// It has a constant colour, constant mass, and variable speed.
// The colour is stored as an index of some color.palette
type Pixicle struct {
	// Colour, contains the index of some palette
	Colour int
	// Velocity is the 2d speed that the particle currently has
	Velocity geom.Vec
	// Mass is the constant movement inertia value of this Pixicle.
	Mass float64
	// Position stores the high-accuracy location of the Pixicle.
	Position    geom.Vec
	newPosition geom.Vec
	newVelocity geom.Vec
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
	Palette color.Palette
	// Calculation is a function that calculates the new position and
	// velocity of a pixicle, given the set of all pixicles in the image
	Calculation   func(pix *Pixicle, pixs *[]Pixicle, timestep float64, options map[string]any)
	width, height int
	options       map[string]any
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

// calculate calls the calculation function on all pixicles
func (p Particled) calculate(timestep float64) {
	for _, pixicle := range *p.Pixicles {
		p.Calculation(&pixicle, p.Pixicles, timestep, p.options)
	}
}

// update updates all pixicles to the new positions and velocities
func (p Particled) update() {
	for _, pixicle := range *p.Pixicles {
		pixicle.Position = pixicle.newPosition
		pixicle.Velocity = pixicle.newVelocity
	}
}

// Iterate runs through one timestep of the physics loop.
func (p Particled) Iterate(timestep float64) {
	p.calculate(timestep)
	p.update()
}

func squareDist(p1, p2 *Pixicle) float64 {
	tmp := math.Pow((p1.Position[0] - p2.Position[0]), 2)
	tmp += math.Pow((p1.Position[1] - p2.Position[1]), 2)

	return tmp
}

// force along axis between two points, positive if attraction
func gravityForce(p1, p2 *Pixicle, likeness float64) geom.Vec {
	direction := p2.Position.Sub(&p1.Position)
	var force float64
	dist := squareDist(p1, p2)
	if dist > 0 {
		force = likeness * p1.Mass * p2.Mass / squareDist(p1, p2)
	}

	return direction.Scale(force)
}

func totalGravityForce(pix *Pixicle, pixs *[]Pixicle, options map[string]any) geom.Vec {
	var force geom.Vec
	var likeness float64
	var currentForce geom.Vec
	// current implementation is very naive Euler...
	for _, other := range *pixs {
		likeness = options["likeness"].(func(int, int) float64)(pix.Colour, other.Colour)
		currentForce = gravityForce(pix, &other, likeness)
		force = force.Add(&currentForce)
	}

	return force
}

func eulerMethod(pix *Pixicle, force geom.Vec, timestep float64) {
	deltaPosition := pix.Velocity.Scale(timestep)
	pix.newPosition = pix.Position.Add(&deltaPosition)

	deltaVelocity := force.Scale(timestep)
	pix.newVelocity = pix.Velocity.Add(&deltaVelocity)
}

// GravityCalculation performs the simple gravity equation to one pixicle.
// The options parameter contains the keys ..., which map to values ...:
//   - "likeness" : func(i,j int) float64 : returns likeness between two colourIndexes.
//   - "..."
func GravityCalculation(pix *Pixicle, pixs *[]Pixicle, timestep float64, options map[string]any) {
	// TODO: the RK4 implementation!

	force := totalGravityForce(pix, pixs, options)

	eulerMethod(pix, force, timestep)
}
