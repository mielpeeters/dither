# dither
Implements a simple Floyd Steinberg Dithering program.

## Install
Run `go build`, an executable file `process` is created.

## Colours
There are 2 ways to define the colours that the algorithm is alowed to choose from:
1. By setting the flag `-colors` and adding as a value a name of a colorpalette that's stored inside the `colorpalette.json` file
2. By not setting any flag, or setting `-k` to define the amount of colours. A k-means algorihtm will run to determine the most important colours in the image.

## Scale
By setting `-scale` to a custom value, the image gets downscaled by the inputted factor.

## More configuration
Running the command below, all options will be displayed.
```shell
./process -help
```
