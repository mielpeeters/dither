module github.com/mielpeeters/dither

go 1.19

replace github.com/mielpeeters/dither => ../dither

require golang.org/x/image v0.6.0

require (
	github.com/kyroy/kdtree v0.0.0-20200419114247-70830f883f1d
	github.com/mielpeeters/pacebar v1.0.4
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
)

require github.com/kyroy/priority-queue v0.0.0-20180327160706-6e21825e7e0c // indirect

replace github.com/mielpeeters/pacebar => ../pacebar
