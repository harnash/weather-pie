package epd

import (
	"image"
	"image/color"

	"github.com/MaxHalford/halfgone"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var blankBuffer []byte

func GetBlankBuffer(logger *zap.SugaredLogger, bounds image.Rectangle) ([]byte, error) {
	if blankBuffer != nil {
		return blankBuffer, nil
	}

	img := image.NewGray(bounds)
	blankBuffer, err := GetBuffer(logger, img, bounds, false)
	return blankBuffer, err
}

func GetBuffer(logger *zap.SugaredLogger, img image.Image, deviceBounds image.Rectangle, dither bool) ([]byte, error) {
	buff := make([]byte, (deviceBounds.Dx()/8)*deviceBounds.Dy())
	for i := 0; i < len(buff); i++ {
		buff[i] = 0x00
	}

	imageBounds := img.Bounds()
	grayImage := image.NewGray(imageBounds)
	for y := 0; y < imageBounds.Max.Y; y++ {
		for x := 0; x < imageBounds.Max.X; x++ {
			grayImage.Set(x, y, color.GrayModel.Convert(img.At(x, y)))
		}
	}

	finalImage := grayImage
	if dither {
		finalImage = halfgone.ThresholdDitherer{Threshold: 127}.Apply(grayImage)
	}

	if imageBounds == deviceBounds {
		logger.Debug("displaying in vertical mode")
		for y := 0; y < imageBounds.Max.Y; y++ {
			for x := 0; x < imageBounds.Max.X; x++ {
				pos := (x + y*deviceBounds.Dx()) / 8
				if pos >= len(buff) {
					continue
				}
				pix := finalImage.GrayAt(x, y)
				if pix.Y > 0 {
					buff[pos] |= 0x80 >> (uint(x) % uint(8))
				}
			}
		}
	} else if imageBounds.Dx() == deviceBounds.Dy() && imageBounds.Dy() == deviceBounds.Dx() {
		logger.Debug("displaying in horizontal")
		for y := 0; y < imageBounds.Max.Y; y++ {
			for x := 0; x < imageBounds.Max.X; x++ {
				newX := y
				newY := deviceBounds.Dy() - x - 1

				pos := (newX + newY*deviceBounds.Dx()) / 8
				if pos >= len(buff) {
					continue
				}
				pix := finalImage.GrayAt(x, y)
				if pix.Y > 0 {
					buff[pos] |= 0x80 >> (uint(y) % uint(8))
				}
			}
		}
	} else {
		return nil, errors.New("invalid image dimensions")
	}

	return buff, nil
}
