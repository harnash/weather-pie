package ui

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/pkg/errors"

	"github.com/BurntSushi/graphics-go/graphics"
)

func RotateImage(src draw.Image) (draw.Image, error) {
	rotImage := image.NewPaletted(src.Bounds(), color.Palette{color.White, color.Black})
	err := graphics.Rotate(rotImage, src, &graphics.RotateOptions{Angle: math.Pi})
	if err != nil {
		return nil, errors.Wrap(err, "could not rotate image")
	}

	return rotImage, nil
}
