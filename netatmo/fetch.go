package netatmo

import (
	"context"
	"strings"
	"time"
	"weather-pi/internal"

	"github.com/spf13/viper"

	"golang.org/x/oauth2"

	"github.com/hekmon/go-netatmo/weather"

	"github.com/hekmon/go-netatmo"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Reading struct {
	Name        string
	Timestamp   time.Time
	Temperature float64
	MinTemp     float64
	MaxTemp     float64
	Humidity    int64
}
type Measurement struct {
	ModuleReadings []Reading
	StationReading *Reading
}

type ModuleInfo struct {
	DeviceId string
	ModuleId string
}

func FetchData(logger *zap.SugaredLogger, sources []internal.Source, apiClientId, apiSecret, token, refreshToken string, since time.Time) ([]Measurement, error) {
	if len(apiClientId) == 0 {
		return nil, errors.New("empty API client ID")
	}
	if len(apiSecret) == 0 {
		return nil, errors.New("empty API secret")
	}
	if len(token) == 0 {
		return nil, errors.New("empty token")
	}
	if len(refreshToken) == 0 {
		return nil, errors.New("empty refreshToken")
	}
	if len(sources) == 0 {
		return nil, errors.New("no measurements to fetch")
	}

	logger.With("clientId", apiClientId).Info("connecting to the Netatmo API")
	oauthBaseConfig := netatmo.OAuth2BaseConfig{
		ClientID:     apiClientId,
		ClientSecret: apiSecret,
		Scopes:       []string{"read_station"},
	}
	oauthConfig := netatmo.GenerateOAuth2Config(oauthBaseConfig)
	oAuthTokens := &oauth2.Token{AccessToken: token, RefreshToken: refreshToken}
	authedClient, err := netatmo.NewClientWithTokens(context.TODO(), oauthConfig, oAuthTokens, nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not connect to the Netatmo API")
	}

	tokens := authedClient.GetTokens()
	viper.Set("token", tokens.AccessToken)
	viper.Set("refreshToken", tokens.RefreshToken)
	if err = viper.WriteConfig(); err != nil {
		logger.With("error", err).Error("could not save generated OAuth tokens")
	}

	logger.Info("fetching stations data")
	client := weather.New(authedClient)
	devices, _, _, err := client.GetStationData(context.TODO(), weather.GetStationDataParameters{})
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch data from the Netatmo API")
	}
	logger.With("num_devices", len(devices.Devices)).Debug("got response with stations data")

	foundMeasurements := 0
	now := time.Now().UTC()
	var measurements []Measurement
	for _, device := range devices.Devices {
		for _, source := range sources {
			log := logger.With("home_name", device.HomeName)
			log.Debug("found station name")
			if device.HomeName == source.StationName {
				log = log.With("station_name", device.ModuleName, "device_id", device.ID)
				log.Info("found station with a proper name")
				data := Measurement{ModuleReadings: []Reading{}}
				data.StationReading = &Reading{
					Name:        device.ModuleName,
					Temperature: device.DashboardData.Temperature,
					MinTemp:     device.DashboardData.TempMin,
					MaxTemp:     device.DashboardData.TempMax,
					Humidity:    int64(device.DashboardData.Humidity),
					Timestamp:   device.DashboardData.Time,
				}

				for _, module := range device.Modules {
					log.With("module_name", module.ModuleName, "configured_names", source.ModuleNames).Debug("found module name")
					for _, moduleName := range source.ModuleNames {
						if strings.TrimSpace(moduleName) == strings.TrimSpace(module.ModuleName) {
							log.With("since", since.Unix(), "until", now.Unix()).Info("found module with a proper name - fetching data")
							log.With("module", module).Info("found module")
							if module.DashboardDataIndoor != nil {
								data.ModuleReadings = append(data.ModuleReadings, Reading{
									Name:        module.ModuleName,
									Temperature: module.DashboardDataIndoor.Temperature,
									MinTemp:     module.DashboardDataIndoor.MinTemp,
									MaxTemp:     module.DashboardDataIndoor.MaxTemp,
									Humidity:    module.DashboardDataIndoor.Humidity,
									Timestamp:   module.DashboardDataIndoor.Time,
								})
								foundMeasurements++
							} else if module.DashboardDataOutdoor != nil {
								data.ModuleReadings = append(data.ModuleReadings, Reading{
									Name:        module.ModuleName,
									Temperature: module.DashboardDataOutdoor.Temperature,
									MinTemp:     module.DashboardDataOutdoor.MinTemp,
									MaxTemp:     module.DashboardDataOutdoor.MaxTemp,
									Humidity:    module.DashboardDataOutdoor.Humidity,
									Timestamp:   module.DashboardDataOutdoor.Time,
								})
								foundMeasurements++
							}
						}
					}
				}
				measurements = append(measurements, data)
			}
		}
	}
	logger.With("num", foundMeasurements).Info("finished fetching measurement data")

	return measurements, nil
}
