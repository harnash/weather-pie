package netatmo

import (
	"context"
	"time"
	"weather-pi/internal"

	"github.com/mikan/netatmo-weather-go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Reading struct {
	Name        string
	Timestamp   time.Time
	Temperature *float64
	MinTemp     *float64
	MaxTemp     *float64
	Humidity    *int
}
type Measurement struct {
	ModuleReadings []Reading
	StationReading *Reading
}

type ModuleInfo struct {
	DeviceId string
	ModuleId string
}

func FetchData(logger *zap.SugaredLogger, sources []internal.Source, apiClientId, apiSecret, username, password string) ([]Measurement, error) {
	if len(apiClientId) == 0 {
		return nil, errors.New("empty API client ID")
	}
	if len(apiSecret) == 0 {
		return nil, errors.New("empty API secret")
	}
	if len(username) == 0 {
		return nil, errors.New("empty username")
	}
	if len(password) == 0 {
		return nil, errors.New("empty password")
	}
	if len(sources) == 0 {
		return nil, errors.New("no measurements to fetch")
	}

	logger.With("clientId", apiClientId).Info("connecting to the Netatmo API")
	client, err := netatmo.NewClient(
		context.Background(),
		apiClientId,
		apiSecret,
		username,
		password)
	if err != nil {
		return nil, errors.Wrap(err, "could not connect to the Netatmo API")
	}

	logger.Info("fetching stations data")
	devices, _, err := client.GetStationsData()
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch data from the Netatmo API")
	}
	logger.With("num_devices", len(devices)).Debug("got response with stations data")

	foundMeasurements := 0
	var measurements []Measurement
	for _, device := range devices {
		for _, source := range sources {
			logger.With("station_name", device.StationName).Debug("found station name")
			if device.StationName == source.StationName {
				logger.With("station_name", device.StationName).Info("found station with a proper name")
				data := Measurement{ModuleReadings: []Reading{}}
				data.StationReading = &Reading{
					Name:        device.StationName,
					Temperature: device.DashboardData.Temperature,
					MinTemp:     device.DashboardData.MinTemperature,
					MaxTemp:     device.DashboardData.MinTemperature,
					Humidity:    device.DashboardData.Humidity,
					Timestamp:   time.Unix(device.DashboardData.UTCTime, 0),
				}

				for _, module := range device.Modules {
					logger.With("module_name", module.ModuleName).Debug("found module name")
					for _, moduleName := range source.ModuleNames {
						if moduleName == module.ModuleName {
							logger.With("module_name", module.ModuleName).Info("found module with a proper name - fetching data")
							data.ModuleReadings = append(data.ModuleReadings, Reading{
								Name:        module.ModuleName,
								Temperature: module.DashboardData.Temperature,
								MinTemp:     module.DashboardData.MinTemperature,
								MaxTemp:     module.DashboardData.MinTemperature,
								Humidity:    module.DashboardData.Humidity,
								Timestamp:   time.Unix(module.DashboardData.UTCTime, 0)})
							foundMeasurements++
						}
					}
				}
				measurements = append(measurements, data)
			}
		}
	}
	logger.With("found_measurements", foundMeasurements).Info("finished fetching measurement data")

	return measurements, nil
}
