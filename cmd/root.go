/*
Package cmd
Copyright © 2021 Łukasz Harasimowicz <dev@harnash.eu>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"image/png"
	"os"
	"time"
	"weather-pi/epd"
	"weather-pi/internal"
	"weather-pi/netatmo"
	"weather-pi/ui"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string
var appConfig internal.Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "weather-pie",
	Short: "raspberry pi netatmo monitor station",
	Long: `Program that will fetch information from the
Netatmo weather station and display current info on e-Paper
display connected to a raspberry pi.`,
	Run: RunApp,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.weather-pie.yaml)")
	rootCmd.PersistentFlags().String("clientId", "", "client ID used to connect to the Netatmo API")
	rootCmd.PersistentFlags().String("secret", "", "secret used to connect to the Netatmo API")
	rootCmd.PersistentFlags().String("username", "", "username of the Netatmo account")
	rootCmd.PersistentFlags().String("password", "", "password of the Netatmo account")
	rootCmd.PersistentFlags().Bool("testMode", false, "run the app in test mode (output test image without connecting to a device")
	rootCmd.PersistentFlags().String("logLevel", "info", "logger log level")
	rootCmd.PersistentFlags().Bool("rotate180", false, "should image be rotated 180 degrees")
	rootCmd.PersistentFlags().Duration("timeWindow", 2*time.Hour, "how large would be the time window to fetch measurements (min/max)")

	if err := viper.BindPFlag("clientId", rootCmd.PersistentFlags().Lookup("clientId")); err != nil {
		zap.S().With("err", err, "flag", "clientId").Fatal("could not bind flag to a config variable")
	}
	if err := viper.BindPFlag("secret", rootCmd.PersistentFlags().Lookup("secret")); err != nil {
		zap.S().With("err", err, "flag", "secret").Fatal("could not bind flag to a config variable")
	}
	if err := viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username")); err != nil {
		zap.S().With("err", err, "flag", "username").Fatal("could not bind flag to a config variable")
	}
	if err := viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password")); err != nil {
		zap.S().With("err", err, "flag", "password").Fatal("could not bind flag to a config variable")
	}
	if err := viper.BindPFlag("logLevel", rootCmd.PersistentFlags().Lookup("logLevel")); err != nil {
		zap.S().With("err", err, "flag", "logLevel").Fatal("could not bind flag to a config variable")
	}
	if err := viper.BindPFlag("testMode", rootCmd.PersistentFlags().Lookup("testMode")); err != nil {
		zap.S().With("err", err, "flag", "testMode").Fatal("could not bind flag to a config variable")
	}
	if err := viper.BindPFlag("rotate180", rootCmd.PersistentFlags().Lookup("rotate180")); err != nil {
		zap.S().With("err", err, "flag", "rotate180").Fatal("could not bind flag to a config variable")
	}
	if err := viper.BindPFlag("timeWindow", rootCmd.PersistentFlags().Lookup("timeWindow")); err != nil {
		zap.S().With("err", err, "flag", "timeWindow").Fatal("could not bind flag to a config variable")
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".weather-pie" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".weather-pie")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	if err := viper.Unmarshal(&appConfig); err != nil {
		fmt.Println("could not unmarshal config")
		os.Exit(2)
	}
}

func RunApp(cmd *cobra.Command, args []string) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logLevel := zap.NewAtomicLevel()
	config.Level = logLevel
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	logger, _ := config.Build()
	sugaredLogger := logger.Sugar()
	if err := logLevel.UnmarshalText([]byte(appConfig.LogLevel)); err != nil {
		logLevel.SetLevel(zap.InfoLevel)
	}

	tm := time.Now().UTC().Add(-appConfig.TimeWindow)
	data, err := netatmo.FetchData(sugaredLogger, appConfig.Sources, appConfig.ClientId, appConfig.ClientSecret, appConfig.Username, appConfig.Password, tm)
	if err != nil {
		sugaredLogger.With("err", err).Error("could not fetch data")
		os.Exit(3)
	}

	e := epd.NewEpd2in13v3(sugaredLogger)
	bImage, rImage, err := ui.BuildGUI(sugaredLogger, e.BoundsHorizontal(), data)
	if err != nil {
		sugaredLogger.With("err", err).Error("could not generate UI")
		os.Exit(4)
	}

	if appConfig.Rotate180 {
		bImage, err = ui.RotateImage(bImage)
		if err != nil {
			sugaredLogger.With("err", err).Fatal("could not rotate black image")
		}

		rImage, err = ui.RotateImage(rImage)
		if err != nil {
			sugaredLogger.With("err", err).Fatal("could not rotate red image")
		}
	}

	if appConfig.TestMode {
		bFile, err := os.OpenFile("out_test_b.png", os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			sugaredLogger.With("err", err).Error("could not open test file for write")
			os.Exit(5)
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				sugaredLogger.With("err", err).Error("could not close a file")
			}
		}(bFile)
		err = png.Encode(bFile, bImage)
		if err != nil {
			sugaredLogger.With("err", err).Error("could not encode the output file")
		}

		rFile, err := os.OpenFile("out_test_r.png", os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			sugaredLogger.With("err", err).Error("could not open test file for write")
			os.Exit(5)
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				sugaredLogger.With("err", err).Error("could not close a file")
			}
		}(rFile)
		err = png.Encode(rFile, rImage)
		if err != nil {
			sugaredLogger.With("err", err).Error("could not encode the output file")
		}
	} else {
		defer func(e *epd.Dev2in13v3) {
			if err := e.Close(); err != nil {
				sugaredLogger.With("err", err).Error("could not close device")
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

		bBuff, err := epd.GetBuffer(sugaredLogger, bImage, e.Bounds(), false)
		if err != nil {
			sugaredLogger.With("err", err).Fatal("could not generate buffer for black the GUI image")
		}
		rBuff, err := epd.GetBuffer(sugaredLogger, rImage, e.Bounds(), false)
		if err != nil {
			sugaredLogger.With("err", err).Fatal("could not generate buffer for red the GUI image")
		}
		err = e.Display(bBuff, rBuff)
		if err != nil {
			sugaredLogger.With("err", err).Fatal("could not display GUI")
		}
	}
}
