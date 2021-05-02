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
  "context"
  "fmt"
  "github.com/mikan/netatmo-weather-go"
  "github.com/spf13/cobra"
  "go.uber.org/zap"
  "go.uber.org/zap/zapcore"
  "os"

  homedir "github.com/mitchellh/go-homedir"
  "github.com/spf13/viper"
)


var cfgFile string

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
  rootCmd.PersistentFlags().String("logLevel", "info", "logger log level")

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
}

func RunApp(cmd *cobra.Command, args []string) {
  config := zap.NewDevelopmentConfig()
  config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
  logLevel := zap.NewAtomicLevel()
  config.Level = logLevel
  logger, _ := config.Build()
  sugaredLogger := logger.Sugar()
  if err := logLevel.UnmarshalText([]byte(viper.GetString("lovLevel"))); err != nil {
    logLevel.SetLevel(zap.InfoLevel)
  }

  client, err := netatmo.NewClient(
    context.Background(),
    viper.GetString("clientId"),
    viper.GetString("secret"),
    viper.GetString("username"),
    viper.GetString("password"))
  if err != nil {
    sugaredLogger.With("err", err).Fatal("could not connect to the Netatmo API")
  }

  device, _, err := client.GetStationsData()

  if err != nil {
    sugaredLogger.With("err", err).Fatal("could not fetch data from the Netatmo API")
  }

  measurement, err := client.GetMeasureByNewest(device[0].ID, device[0].Modules[0].ID)
  if err != nil {
    sugaredLogger.With("err", err).Fatal("could not fetch measurements")
  }

  fmt.Printf("Temperature: %.3f", *measurement.Temperature)
}

