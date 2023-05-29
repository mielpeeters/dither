// Package particled implements a new image type that stores each (original)
// pixel with constant colour at variable location.
// This enables colour-physics systems to be implemented on top.
package particled

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/kyroy/kdtree"
	"github.com/kyroy/kdtree/kdrange"
	"github.com/mielpeeters/dither/geom"
	"github.com/mielpeeters/pacebar"
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

// Calculation is a type alias for a function which calculates the new
// position and velocity of pix, based on pixs, timestep and options
type Calculation func(pix kdtree.Point, pixs *kdtree.KDTree, timestep float64, options map[string]any)

// Particled is a type of image that stores pixels at variable locations.
// Each pixel is represented as a Pixicle.
type Particled struct {
	// Particles is a pointer to a slice of all pixicles
	Pixicles *kdtree.KDTree
	// Palette holds the used colourpalette for this image
	Palette color.Palette
	// Calc is a function that calculates the new position and
	// velocity of a pixicle, given the set of all pixicles in the image
	Calc          Calculation
	width, height int
	Options       map[string]any
	Timestep      float64
	pb            pacebar.Pacebar
}

// Dimensions returns the amount of Dimensions for this pixicle
func (pix *Pixicle) Dimensions() int {
	return 2
}

// Dimension returns the position at dimension number i
func (pix *Pixicle) Dimension(i int) float64 {
	return pix.Position[i]
}

func (pix *Pixicle) toCoordinate() coordinate {
	return coordinate{
		x: int(pix.Position[0]),
		y: int(pix.Position[1]),
	}
}

func (pix *Pixicle) equals(p Pixicle) bool {
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

	// fmt.Println("input", cc)
	// fmt.Println("output", average)

	return average
}

func (cc colourClub) max(palette *color.Palette) color.Color {
	counts := make([]int, len(*palette))
	var colourNo int
	var maxValue int
	var maxIndex int

	for _, colour := range cc.pixicles {
		colourNo = palette.Index(colour)

		counts[colourNo]++

		if counts[colourNo] > maxValue {
			maxValue = counts[colourNo]
			maxIndex = colourNo
		}
	}

	return (*palette)[maxIndex]
}

func (cc *colourClub) add(value color.Color) {
	cc.pixicles = append(cc.pixicles, value)
}

// ColorModel returns the Image's color model.
func (p *Particled) ColorModel() color.Model {
	return p.Palette
}

// Bounds returns the domain for which At can return non-zero color.
func (p *Particled) Bounds() image.Rectangle {
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
func (p *Particled) ToPaletted() *image.Paletted {
	paletted := image.NewPaletted(p.Bounds(), p.Palette)

	colourClubs := make(map[coordinate]*colourClub)

	fullRange := kdrange.New(0, float64(p.width), 0, float64(p.height))
	for _, pixicle := range p.Pixicles.RangeSearch(fullRange) {
		coor := pixicle.(*Pixicle).toCoordinate()
		if colourClubs[coor] == nil {
			colourClubs[coor] = &colourClub{
				pixicles: []color.Color{},
			}
		}
		(colourClubs[coor]).add(p.Palette[pixicle.(*Pixicle).Colour])
	}

	for x := 0; x <= p.width; x++ {
		for y := 0; y <= p.height; y++ {
			if (colourClubs[coordinate{x: x, y: y}] == nil) {
				paletted.Set(x, y, p.Palette[0])
			} else {
				paletted.Set(x, y, colourClubs[coordinate{x: x, y: y}].max(&p.Palette))
			}
		}
	}

	return paletted
}

// calculate calls the calculation function on all pixicles
func (p *Particled) calculate() {
	fullRange := kdrange.New(-100000, 100000, -1000000, 100000)
	for _, pixicle := range p.Pixicles.RangeSearch(fullRange) {
		p.Calc(pixicle, p.Pixicles, p.Timestep, p.Options)
		p.pb.Done(1)
	}
}

// update updates all pixicles to the new positions and velocities
func (p *Particled) update() {
	fullRange := kdrange.New(0, float64(p.width), 0, float64(p.height))
	for _, pixicle := range p.Pixicles.RangeSearch(fullRange) {
		pixicle.(*Pixicle).Position = pixicle.(*Pixicle).newPosition
		pixicle.(*Pixicle).Velocity = pixicle.(*Pixicle).newVelocity
	}
}

// Iterate runs through one timestep of the physics loop.
func (p *Particled) Iterate() {
	p.calculate()
	p.update()
	p.Pixicles.Balance()
}

func squareDist(p1, p2 *Pixicle) float64 {
	tmp := math.Pow((p1.Position[0] - p2.Position[0]), 2)
	tmp += math.Pow((p1.Position[1] - p2.Position[1]), 2)

	return tmp
}

// force along axis between two points, positive if attraction
// when pixicles get closer than 1, they are always repelled!
func gravityForce(p1, p2 *Pixicle, likeness float64) geom.Vec {
	direction := p2.Position.Sub(&p1.Position)
	var force float64
	dist := squareDist(p1, p2)

	if dist > 0 {
		if dist < 0.05 {
			force = 0
		} else if dist < 1 {
			force = -p1.Mass * p2.Mass / dist // repelling force
		} else {
			force = likeness * p1.Mass * p2.Mass / dist
		}
	}

	return direction.Scale(force)
}

func totalGravityForce(pix kdtree.Point, pixs *kdtree.KDTree, options map[string]any) geom.Vec {
	var force geom.Vec
	var likeness float64
	var currentForce geom.Vec
	// current implementation is very naive Euler...

	for _, other := range pixs.RangeSearch(kdrange.Range{{pix.Dimension(0) - 5, pix.Dimension(0) + 5}, {pix.Dimension(1) - 5, pix.Dimension(1) + 5}}) {
		likeness = options["likeness"].(func(int, int) float64)(pix.(*Pixicle).Colour, other.(*Pixicle).Colour)
		currentForce = gravityForce(pix.(*Pixicle), other.(*Pixicle), likeness)
		force = force.Add(&currentForce)
	}

	return force
}

// eulerMethod uses velocity and force to set new position and velocity
func eulerMethod(pix kdtree.Point, force geom.Vec, timestep, damping float64) {
	px := pix.(*Pixicle)

	deltaPosition := px.Velocity.Scale(timestep)
	px.newPosition = px.Position.Add(&deltaPosition)

	deltaVelocity := force.Scale(timestep)
	// var sign int = 1
	// if deltaVelocity[0] < 0 {
	// 	sign = -1
	// }
	// deltaVelocity = geom.Vec{float64(sign) * math.Pow(math.Abs(deltaVelocity[0]), (1.0-damping)), deltaVelocity[1]}
	px.newVelocity = px.Velocity.Scale(0.80)
	px.newVelocity = px.newVelocity.Add(&deltaVelocity)

}

// GravityCalculation performs the simple gravity equation to one pixicle.
// The options parameter contains the keys ..., which map to values ...:
//   - "likeness" : func(i,j int) float64 : returns likeness between two colourIndexes.
//   - "..."
func GravityCalculation(pix kdtree.Point, pixs *kdtree.KDTree, timestep float64, options map[string]any) {
	// TODO: the RK4 implementation!

	force := totalGravityForce(pix, pixs, options)

	eulerMethod(pix, force, timestep, 0.0)
}

// sortforce applies a force towards a region corresponding with the pixel colour index
func sortForce(pix *Pixicle, width int, k int) geom.Vec {
	var goal float64
	step := float64(width) / float64(k)

	goal = step/2.0 + step*float64(pix.Colour)

	return geom.Vec{(goal - pix.Position[0]), 0.0}
}

// SortCalculation tempts to sort the pixels horizontally
// options["width"] -> width of the particled image
// options["k"] -> amount of colours
func SortCalculation(pix *Pixicle, pixs []*Pixicle, timestep float64, options map[string]any) {
	force := sortForce(pix, options["width"].(int), options["k"].(int))

	eulerMethod(pix, force, timestep, 0.2)
}

// FromPaletted creates a particled image from a paletted image
func FromPaletted(paletted *image.Paletted, Calc Calculation, timestep float64, options map[string]any) *Particled {
	X := paletted.Rect.Dx()
	Y := paletted.Rect.Dy()
	pixs := make([]kdtree.Point, X*Y)

	for x := 0; x < X; x++ {
		for y := 0; y < Y; y++ {
			pixs[x+X*y] = &Pixicle{
				Colour:      int(paletted.ColorIndexAt(x, y)),
				Velocity:    [2]float64{0.0, 0.0},
				Mass:        1,
				Position:    [2]float64{float64(x), float64(y)},
				newPosition: [2]float64{0.0, 0.0},
				newVelocity: [2]float64{0.0, 0.0},
				id:          x + y*X,
			}
		}
	}

	// creating the KD tree
	kd := kdtree.New(pixs)

	return &Particled{
		Pixicles: kd,
		Palette:  paletted.Palette,
		Calc:     Calc,
		width:    X,
		height:   Y,
		Options:  options,
		Timestep: timestep,
	}
}

// Simulate creates a slice of paletted frames using the particled starting point
func (p Particled) Simulate(length int) []*image.Paletted {
	frames := make([]*image.Paletted, length)

	pb := pacebar.Pacebar{
		Work: p.width * p.height * length,
		Name: "Simulation",
	}

	p.pb = pb

	for i := 0; i < length; i++ {
		p.Iterate()
		frames[i] = p.ToPaletted()
	}

	return frames
}
