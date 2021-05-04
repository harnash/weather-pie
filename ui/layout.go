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

func BuildGUI(logger *zap.SugaredLogger, bounds image.Rectangle, measurement []netatmo.Measurement) (blackImg draw.Image, redImg draw.Image, err error) {
	if len(measurement) != 1 {
		err = errors.New("measurements incomplete")
		return
	}
	fontData, err := freetype.ParseFont(goregular.TTF)
	if err != nil {
		err = errors.Wrap(err, "could not parse font file")
		return
	}

	blackImg = image.NewGray(bounds)
	draw.Draw(blackImg, blackImg.Bounds(), image.White, image.Point{}, draw.Src)
	fontCtx := freetype.NewContext()
	fontCtx.SetFont(fontData)
	fontCtx.SetFontSize(20)
	fontCtx.SetDPI(96)
	fontCtx.SetClip(blackImg.Bounds())
	fontCtx.SetDst(blackImg)
	fontCtx.SetSrc(image.Black)
	fontCtx.SetHinting(font.HintingFull)

	// Temperatures
	pt := freetype.Pt(10, 15+int(fontCtx.PointToFixed(20)>>6))
	_, err = fontCtx.DrawString(fmt.Sprintf("%.1f °C", *measurement[0].ModuleReadings[0].Temperature), pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw 1st temperature string")
	}

	pt = freetype.Pt(110, 15+int(fontCtx.PointToFixed(20)>>6))
	_, err = fontCtx.DrawString(fmt.Sprintf("%.1f °C", *measurement[0].StationReading.Temperature), pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw 2nd temperature string")
	}

	// Humidity
	fontCtx.SetFontSize(14)
	pt = freetype.Pt(10, 40+int(fontCtx.PointToFixed(14)>>6))
	_, err = fontCtx.DrawString(fmt.Sprintf("%d%%", *measurement[0].ModuleReadings[0].Humidity), pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw 1st humidity string")
	}

	pt = freetype.Pt(110, 40+int(fontCtx.PointToFixed(14)>>6))
	_, err = fontCtx.DrawString(fmt.Sprintf("%d%%", *measurement[0].StationReading.Humidity), pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw 2nd humidity string")
	}

	redImg = image.NewGray(bounds)
	draw.Draw(redImg, redImg.Bounds(), image.White, image.Point{}, draw.Src)

	// Names
	fontCtx2 := freetype.NewContext()
	fontCtx2.SetFont(fontData)
	fontCtx2.SetFontSize(10)
	fontCtx2.SetDPI(96)
	fontCtx2.SetClip(redImg.Bounds())
	fontCtx2.SetDst(redImg)
	fontCtx2.SetSrc(image.Black)
	fontCtx2.SetHinting(font.HintingFull)

	pt = freetype.Pt(10, 1+int(fontCtx2.PointToFixed(10)>>6))
	_, err = fontCtx2.DrawString(measurement[0].StationReading.Name, pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw station name string")
	}

	fontCtx2.SetFontSize(10)
	pt = freetype.Pt(110, 1+int(fontCtx2.PointToFixed(10)>>6))
	_, err = fontCtx2.DrawString(measurement[0].ModuleReadings[0].Name, pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw module name string")
	}

	return
}
