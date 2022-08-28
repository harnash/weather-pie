package netatmo

import (
	"context"
	"strings"
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

func FetchData(logger *zap.SugaredLogger, sources []internal.Source, apiClientId, apiSecret, username, password string, since time.Time) ([]Measurement, error) {
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
	now := time.Now().UTC()
	var measurements []Measurement
	for _, device := range devices {
		for _, source := range sources {
			log := logger.With("station_name", device.StationName)
			log.Debug("found station name")
			if device.StationName == source.StationName {
				log = log.With("station_name", device.StationName, "device_id", device.ID)
				log.Info("found station with a proper name")
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
					log.With("module_name", module.ModuleName, "configured_names", source.ModuleNames).Debug("found module name")
					for _, moduleName := range source.ModuleNames {
						if strings.TrimSpace(moduleName) == strings.TrimSpace(module.ModuleName) {
							log.With("since", since.Unix(), "until", now.Unix()).Info("found module with a proper name - fetching data")
							measurements, err := client.GetMeasureByTimeRange(device.ID, module.ID, since.Unix(), now.Unix())
							if err != nil || len(measurements) == 0 {
								logger.With("error", err).Error("could not fetch measurements")
								continue
							}
							log.With("num", len(measurements)).Info("fetched measurements")
							var minTemp *float64
							var maxTemp *float64
							for idx, measure := range measurements {
								if measure.Temperature == nil {
									continue
								}

								if idx == 0 || minTemp == nil || *minTemp > *measure.Temperature {
									minTemp = measure.Temperature
								}
								if idx == 0 || maxTemp == nil || *maxTemp < *measure.Temperature {
									maxTemp = measure.Temperature
								}
								if idx == len(measurements)-1 {
									data.ModuleReadings = append(data.ModuleReadings, Reading{
										Name:        module.ModuleName,
										Temperature: measure.Temperature,
										MinTemp:     minTemp,
										MaxTemp:     maxTemp,
										Humidity:    module.DashboardData.Humidity,
										Timestamp:   time.Unix(module.LastSeenTime, 0).Local(),
									})
									foundMeasurements++
								}
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
