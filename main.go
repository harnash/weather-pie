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
	"time"
	"weather-pi/epd"
)

//go:embed resources/2in13bc-b.png
//go:embed resources/2in13bc-ry.png
var resources embed.FS

func main() {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := config.Build()
	sugaredLogger := logger.Sugar()

	e := epd.NewEpd2in13v3(sugaredLogger)
	defer func(e *epd.Dev2in13v3) {
		if err := e.Close(); err != nil {
			sugaredLogger.Error("could not close device", "err", err)
		}
	}(e)
	defer func(e *epd.Dev2in13v3) {
		if err := e.Clear(); err != nil {
			sugaredLogger.Error("could not clear the device", "err", err)
		}
	}(e)
	err := e.Init()
	if err != nil {
		sugaredLogger.Fatal("error while initializing device", "err", err)
	}

	err = e.Clear()
	if err != nil {
		sugaredLogger.Fatal("error while clearing the device screen", "err", err)
	}

	sugaredLogger.Info("displaying images: 2in13bc-b.png and 2in13bc-ry.png")
	fileBlack, err := resources.Open("resources/2in13bc-b.png")
	if err != nil {
		sugaredLogger.Fatalf("error opening image: %s\n", err)
	}
	defer func(fileBlack fs.File) {
		if err := fileBlack.Close(); err != nil {
			sugaredLogger.Error("could not close a file", "err", err)
		}
	}(fileBlack)

	blackImage, _, err := image.Decode(fileBlack)
	if err != nil {
		sugaredLogger.Fatal("error decoding image", "err", err)
	}

	fileRed, err := resources.Open("resources/2in13bc-ry.png")
	if err != nil {
		sugaredLogger.Fatal("error opening image", "err", err)
	}
	defer func(fileRed fs.File) {
		if err := fileRed.Close(); err != nil {
			sugaredLogger.Error("could not close a file", "err", err)
		}
	}(fileRed)

	redImage, _, err := image.Decode(fileRed)
	if err != nil {
		sugaredLogger.Fatal("error decoding image", "err", err)
	}

	blackBuffer, err := epd.GetBuffer(sugaredLogger, blackImage, e.Bounds(),false)
	if err != nil {
		sugaredLogger.Fatal("could not generate buffer for black image", "err", err)
	}
	redBuffer, err := epd.GetBuffer(sugaredLogger, redImage, e.Bounds(),false)
	if err != nil {
		sugaredLogger.Fatal("could not generate buffer for red image", "err", err)
	}
	err = e.Display(blackBuffer, redBuffer)
	if err != nil {
		sugaredLogger.Fatal("error while displaying image", "err", err)
	}

	sugaredLogger.Info("sleeping")
	time.Sleep(5 * time.Second)

	if err := e.Clear(); err != nil {
		sugaredLogger.Fatal("could not clear the device", "err", err)
	}
	fontData, err := freetype.ParseFont(goregular.TTF)
	if err != nil {
		sugaredLogger.Fatal("could not parse font file", "err", err)
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
		sugaredLogger.Fatal("could not draw string", "err", err)
	}

	blankImg := image.NewGray(e.BoundsHorizontal())
	draw.Draw(blankImg, blankImg.Bounds(), image.White, image.Point{}, draw.Src)
	blackBuffer, err = epd.GetBuffer(sugaredLogger, img, e.Bounds(), true)
	if err != nil {
		sugaredLogger.Fatal("could not generate buffer for black text image", "err", err)
	}

	redBuffer, err = epd.GetBuffer(sugaredLogger, blankImg, e.Bounds(), false)
	_ = e.Display(blackBuffer, redBuffer)
	logger.Info("sleeping")
	time.Sleep(5 * time.Second)
}
