package ui

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"time"
	"weather-pi/netatmo"

	"github.com/pkg/errors"

	"github.com/golang/freetype"
	"go.uber.org/zap"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
)

const deviceDPI = 110
const mainFontSize = 18
const secondaryFontSize = 12
const tertiaryFontSize = 8
const statusFontSize = 7

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

	leftPane := image.Rect(1, 1, bounds.Dx()/2-1, bounds.Dy()-1)
	rightPane := image.Rect((bounds.Dx()/2)+1, 1, bounds.Dx()-1, bounds.Dy()-1)

	blackImg = image.NewPaletted(bounds, color.Palette{color.White, color.Black})
	draw.Draw(blackImg, blackImg.Bounds(), image.White, image.Point{}, draw.Src)
	fontCtx := freetype.NewContext()
	fontCtx.SetFont(fontData)
	fontCtx.SetDPI(deviceDPI)
	fontCtx.SetClip(blackImg.Bounds())
	fontCtx.SetDst(blackImg)
	fontCtx.SetSrc(image.Black)
	fontCtx.SetHinting(font.HintingFull)

	// Names
	fontCtx.SetFontSize(tertiaryFontSize)
	pt := freetype.Pt(leftPane.Min.X, 1+int(fontCtx.PointToFixed(tertiaryFontSize)>>6))
	_, err = fontCtx.DrawString(measurement[0].StationReading.Name, pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw station name string")
	}

	pt = freetype.Pt(rightPane.Min.X, 1+int(fontCtx.PointToFixed(tertiaryFontSize)>>6))
	_, err = fontCtx.DrawString(measurement[0].ModuleReadings[0].Name, pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw module name string")
	}

	// Humidity label
	fontCtx.SetFontSize(tertiaryFontSize)
	pt = freetype.Pt(leftPane.Min.X, 72+int(fontCtx.PointToFixed(tertiaryFontSize)>>6))
	_, err = fontCtx.DrawString("H:", pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw 1st humidity label")
	}

	pt = freetype.Pt(rightPane.Min.X, 72+int(fontCtx.PointToFixed(tertiaryFontSize)>>6))
	_, err = fontCtx.DrawString("H:", pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw 2nd humidity label")
	}

	// Temperatures range label
	fontCtx.SetFontSize(statusFontSize)
	pt = freetype.Pt(leftPane.Min.X, 45+int(fontCtx.PointToFixed(statusFontSize)>>6))
	_, err = fontCtx.DrawString("Min:", pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw 1st min temperature label")
	}
	pt = freetype.Pt(leftPane.Min.X+(leftPane.Dx()/2), 45+int(fontCtx.PointToFixed(statusFontSize)>>6))
	_, err = fontCtx.DrawString("Max:", pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw 1st max temperature label")
	}

	pt = freetype.Pt(rightPane.Min.X, 45+int(fontCtx.PointToFixed(statusFontSize)>>6))
	_, err = fontCtx.DrawString("Min:", pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw 2nd min temperature label")
	}
	pt = freetype.Pt(rightPane.Min.X+(rightPane.Dx()/2), 45+int(fontCtx.PointToFixed(statusFontSize)>>6))
	_, err = fontCtx.DrawString("Max:", pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw 2nd max temperature label")
	}

	pt = freetype.Pt(leftPane.Min.X, 90+int(fontCtx.PointToFixed(statusFontSize)>>6))
	_, err = fontCtx.DrawString(fmt.Sprintf("Ts: %s", measurement[0].StationReading.Timestamp.Format(time.ANSIC)), pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw timestamp")
	}

	// Red layer
	redImg = image.NewPaletted(bounds, color.Palette{color.White, color.Black})
	draw.Draw(redImg, redImg.Bounds(), image.White, image.Point{}, draw.Src)

	// Humidity
	fontCtx.SetFontSize(secondaryFontSize)
	pt = freetype.Pt(leftPane.Min.X+15, 70+int(fontCtx.PointToFixed(secondaryFontSize)>>6))
	_, err = fontCtx.DrawString(fmt.Sprintf("%d%%", *measurement[0].ModuleReadings[0].Humidity), pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw 1st humidity string")
	}

	pt = freetype.Pt(rightPane.Min.X+15, 70+int(fontCtx.PointToFixed(secondaryFontSize)>>6))
	_, err = fontCtx.DrawString(fmt.Sprintf("%d%%", *measurement[0].StationReading.Humidity), pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw 2nd humidity string")
	}

	// Temperatures
	fontCtx.SetFontSize(mainFontSize)
	fontCtx.SetDst(redImg)
	fontCtx.SetClip(redImg.Bounds())
	pt = freetype.Pt(leftPane.Min.X, 15+int(fontCtx.PointToFixed(mainFontSize)>>6))
	_, err = fontCtx.DrawString(fmt.Sprintf("%.1f°C", *measurement[0].ModuleReadings[0].Temperature), pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw 1st temperature string")
	}

	pt = freetype.Pt(rightPane.Min.X, 15+int(fontCtx.PointToFixed(mainFontSize)>>6))
	_, err = fontCtx.DrawString(fmt.Sprintf("%.1f°C", *measurement[0].StationReading.Temperature), pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw 2nd temperature string")
	}

	// Temperature ranges
	fontCtx.SetFontSize(tertiaryFontSize)
	pt = freetype.Pt(leftPane.Min.X, 55+int(fontCtx.PointToFixed(tertiaryFontSize)>>6))
	_, err = fontCtx.DrawString(fmt.Sprintf("%.1f°C", *measurement[0].ModuleReadings[0].MinTemp), pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw 1st min temperature string")
	}

	pt = freetype.Pt(leftPane.Min.X+(leftPane.Dx()/2), 55+int(fontCtx.PointToFixed(tertiaryFontSize)>>6))
	_, err = fontCtx.DrawString(fmt.Sprintf("%.1f°C", *measurement[0].ModuleReadings[0].MaxTemp), pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw 1st max temperature string")
	}

	pt = freetype.Pt(rightPane.Min.X, 55+int(fontCtx.PointToFixed(tertiaryFontSize)>>6))
	_, err = fontCtx.DrawString(fmt.Sprintf("%.1f°C", *measurement[0].StationReading.MinTemp), pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw 2nd min temperature string")
	}

	pt = freetype.Pt(rightPane.Min.X+(rightPane.Dx()/2), 55+int(fontCtx.PointToFixed(tertiaryFontSize)>>6))
	_, err = fontCtx.DrawString(fmt.Sprintf("%.1f°C", *measurement[0].StationReading.MaxTemp), pt)
	if err != nil {
		logger.With("err", err).Fatal("could not draw 2nd max temperature string")
	}

	return
}
