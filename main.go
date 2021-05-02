package main

import (
	"embed"
	_ "embed"
	"github.com/golang/freetype"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"image"
	_ "image/png"
	"io/fs"
	"os"
	"time"
	"weather-pi/epd"
)

//go:embed resources/2in13bc-b.png
//go:embed resources/2in13bc-ry.png
var resources embed.FS

func main() {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logLevel := zap.NewAtomicLevel()
	config.Level = logLevel
	logger, _ := config.Build()
	sugaredLogger := logger.Sugar()
	if err := logLevel.UnmarshalText([]byte(os.Getenv("LOG_LEVEL"))); err != nil {
		logLevel.SetLevel(zap.InfoLevel)
	}

	e := epd.NewEpd2in13v3(sugaredLogger)
	defer func(e *epd.Dev2in13v3) {
		if err := e.Close(); err != nil {
			sugaredLogger.With("err", err).Error("could not close device")
		}
	}(e)
	defer func(e *epd.Dev2in13v3) {
		if err := e.Clear(); err != nil {
			sugaredLogger.With("err", err).Error("could not clear the device")
		}
	}(e)
	err := e.Init()
	if err != nil {
		sugaredLogger.With("err", err).Fatal("error while initializing device")
	}

	err = e.Clear()
	if err != nil {
		sugaredLogger.With("err", err).Fatal("error while clearing the device screen")
	}

	sugaredLogger.Info("displaying images: 2in13bc-b.png and 2in13bc-ry.png")
	fileBlack, err := resources.Open("resources/2in13bc-b.png")
	if err != nil {
		sugaredLogger.With("err", err).Fatalf("error opening image")
	}
	defer func(fileBlack fs.File) {
		if err := fileBlack.Close(); err != nil {
			sugaredLogger.With("err", err).Error("could not close a file")
		}
	}(fileBlack)

	blackImage, _, err := image.Decode(fileBlack)
	if err != nil {
		sugaredLogger.With("err", err).Fatal("error decoding image")
	}

	fileRed, err := resources.Open("resources/2in13bc-ry.png")
	if err != nil {
		sugaredLogger.With("err", err).Fatal("error opening image")
	}
	defer func(fileRed fs.File) {
		if err := fileRed.Close(); err != nil {
			sugaredLogger.With("err", err).Error("could not close a file")
		}
	}(fileRed)

	redImage, _, err := image.Decode(fileRed)
	if err != nil {
		sugaredLogger.With("err", err).Fatal("error decoding image")
	}

	blackBuffer, err := epd.GetBuffer(sugaredLogger, blackImage, e.Bounds(),false)
	if err != nil {
		sugaredLogger.With("err", err).Fatal("could not generate buffer for black image")
	}
	redBuffer, err := epd.GetBuffer(sugaredLogger, redImage, e.Bounds(),false)
	if err != nil {
		sugaredLogger.With("err", err).Fatal("could not generate buffer for red image")
	}
	err = e.Display(blackBuffer, redBuffer)
	if err != nil {
		sugaredLogger.With("err", err).Fatal("error while displaying image")
	}

	sugaredLogger.Info("sleeping")
	time.Sleep(5 * time.Second)

	if err := e.Clear(); err != nil {
		sugaredLogger.With("err", err).Fatal("could not clear the device")
	}
	fontData, err := freetype.ParseFont(goregular.TTF)
	if err != nil {
		sugaredLogger.With("err", err).Fatal("could not parse font file")
	}

	img := image.NewGray(e.BoundsHorizontal())
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
	_, err = fontCtx.DrawString("Karola", pt)
	if err != nil {
		sugaredLogger.With("err", err).Fatal("could not draw string")
	}

	blankImg := image.NewGray(e.BoundsHorizontal())
	draw.Draw(blankImg, blankImg.Bounds(), image.White, image.Point{}, draw.Src)
	blackBuffer, err = epd.GetBuffer(sugaredLogger, img, e.Bounds(), true)
	if err != nil {
		sugaredLogger.With("err", err).Fatal("could not generate buffer for black text image")
	}

	redBuffer, err = epd.GetBuffer(sugaredLogger, blankImg, e.Bounds(), false)
	_ = e.Display(blackBuffer, redBuffer)
	logger.Info("sleeping")
	time.Sleep(5 * time.Second)
}
