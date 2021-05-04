package ui

import (
	"fmt"
	"image"
	"image/draw"
	"weather-pi/netatmo"

	"github.com/pkg/errors"

	"github.com/golang/freetype"
	"go.uber.org/zap"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
)

func BuildGUI(logger *zap.SugaredLogger, bounds image.Rectangle, measurement []netatmo.Measurement) (image.Image, error) {
	if len(measurement) != 1 {
		return nil, errors.New("measurements incomplete")
	}
	fontData, err := freetype.ParseFont(goregular.TTF)
	if err != nil {
		logger.With("err", err).Fatal("could not parse font file")
	}

	img := image.NewGray(bounds)
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)
	fontCtx := freetype.NewContext()
	fontCtx.SetFont(fontData)
	fontCtx.SetFontSize(12)
	fontCtx.SetDPI(96)
	fontCtx.SetClip(img.Bounds())
	fontCtx.SetDst(img)
	fontCtx.SetSrc(image.Black)
	fontCtx.SetHinting(font.HintingFull)

	pt := freetype.Pt(10, 10+int(fontCtx.PointToFixed(12)>>6))
	_, err = fontCtx.DrawString(fmt.Sprintf("%.1f °C", *measurement[0].ModuleReadings[0].Temperature), pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw string")
	}

	pt = freetype.Pt(10, 30+int(fontCtx.PointToFixed(12)>>6))
	_, err = fontCtx.DrawString(fmt.Sprintf("%d%%", *measurement[0].ModuleReadings[0].Humidity), pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw string")
	}

	pt = freetype.Pt(100, 10+int(fontCtx.PointToFixed(12)>>6))
	_, err = fontCtx.DrawString(fmt.Sprintf("%.1f °C", *measurement[0].StationReading.Temperature), pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw string")
	}

	pt = freetype.Pt(100, 30+int(fontCtx.PointToFixed(12)>>6))
	_, err = fontCtx.DrawString(fmt.Sprintf("%d%%", *measurement[0].StationReading.Humidity), pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw string")
	}

	return img, nil
}
