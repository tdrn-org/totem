package openmeteo

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	openmeteoapi "github.com/tdrn-org/go-open-meteo/api"
)

func TestOpenMeteoCurrent(t *testing.T) {
	apiURL, err := url.Parse("https://api.open-meteo.com")
	require.NoError(t, err)
	apiClient, err := openmeteoapi.NewClientWithResponses(apiURL.String())
	require.NoError(t, err)
	temperaturUnit := openmeteoapi.Celsius
	windSpeedUnit := openmeteoapi.Kmh
	precipitationUnit := openmeteoapi.Mm
	timeformat := openmeteoapi.Iso8601
	timezone := "Europe/Berlin"
	params := &openmeteoapi.GetV1ForecastParams{
		Latitude:          "48.17",
		Longitude:         "11.57",
		Current:           &paramsCurrent,
		TemperatureUnit:   &temperaturUnit,
		WindSpeedUnit:     &windSpeedUnit,
		PrecipitationUnit: &precipitationUnit,
		Timeformat:        &timeformat,
		Timezone:          &timezone,
	}
	response, err := apiClient.GetV1ForecastWithResponse(t.Context(), params)
	require.NoError(t, err)
	weather, err := currentResponseToCurrentWeather((*openmeteoapi.GetV1Forecast200JSONResponse)(response.JSON200))
	require.NoError(t, err)
	_ = weather
}

func TestOpenMeteo8hWeather(t *testing.T) {
	apiURL, err := url.Parse("https://api.open-meteo.com")
	require.NoError(t, err)
	apiClient, err := openmeteoapi.NewClientWithResponses(apiURL.String())
	require.NoError(t, err)
	temperaturUnit := openmeteoapi.Celsius
	windSpeedUnit := openmeteoapi.Kmh
	precipitationUnit := openmeteoapi.Mm
	timeformat := openmeteoapi.Iso8601
	timezone := "Europe/Berlin"
	forecastDays := 2
	forecastHours := 4
	params := &openmeteoapi.GetV1ForecastParams{
		Latitude:          "48.17",
		Longitude:         "11.57",
		Minutely15:        &paramsMinutely15,
		TemperatureUnit:   &temperaturUnit,
		WindSpeedUnit:     &windSpeedUnit,
		PrecipitationUnit: &precipitationUnit,
		Timeformat:        &timeformat,
		Timezone:          &timezone,
		ForecastDays:      &forecastDays,
		ForecastHours:     &forecastHours,
	}
	response, err := apiClient.GetV1ForecastWithResponse(t.Context(), params)
	require.NoError(t, err)
	weathers, err := minutely15ResponseToWeathers((*openmeteoapi.GetV1Forecast200JSONResponse)(response.JSON200))
	require.NoError(t, err)
	weathers = sortAndFilterWeathers(weathers, 8*time.Hour)
	_ = weathers
}
