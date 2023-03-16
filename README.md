# dither
Implements the Floyd Steinberg Dithering algorithm, and some other image manipulation techniques.
It is a work in progress, more effects are in the work.
This go module contains multiple packages:
- **process**: the core functionality, implementing the image algorithms (like FSD).
- **geom**: some simple geometry types (like `Point`) and functions to manipulate them.
- **imgutil**: some image utilities, like `OpenImage`.
- **kmeans**: k-means clustering implementation, useful for finding the k most prominent colors in an image.
- **nearneigh**: a work in progress, impelementation of a nearest neighbour search algorithm.
- **kdtree**: a work in progress, implements a kd tree search structure for fast nearest neighbour.
- **colorpalette**: a custom defined colorpalette type, with accompanying functions.


<div style="text-align:center"><img src="https://user-images.githubusercontent.com/72082402/225594701-a15c3d26-5ad9-4d42-9d25-cdc7751c8ad2.png" alt="example image created using the dither module." height="400"></div>

## Usage
In your go module (after having run `go mod init mymodule`), get this module by running `go get github.com/mielpeeters/dither`.
You can then import the various packages in your code.

As an example, here is code you could write to apply the Floyd Steinberg Dithering algorithm:
```go
package main

import (
	"github.com/mielpeeters/dither/colorpalette"
	"github.com/mielpeeters/dither/imgutil"
	"github.com/mielpeeters/dither/process"
)

func main() {
    // open the input image
    img, err := imgutil.OpenImage("path/to/imputImage.any")
    if err != nil {
    return
    }

    // convert the input into a slice of slices (2d array) of pixels 
    pixels := imgutil.ImageToPixels(img)

    // scale the image down with a given scale
    scale := 10
    process.Downscale(pixels, scale)

    // get the palette in which to create the new image
    amountOfColors := 5
    sampleFactor := 4
    knnRuns := 3
    palette = colorpalette.Create(pixels, amountOfColors, sampleFactor, knnRuns)

    // apply dithering to the image
    diffusers := process.StuckiDiffuser()
    paletted := process.ApplyErrorDiffusion(pixels, palette, diffusers)

    // save the image as a GIF (efficient for paletted images)
    imgutil.SaveGIF(paletted, "path/to/outputImage.gif")
}
``` 
Note that some of the functions used here are not optimal yet, and the usage will be made simpler in the future.
