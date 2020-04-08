// Package wu is a simple API client for wunderground.com's API.
package wu

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Client struct {
	apiToken  string
	stationID string
}

func New(apiToken string, stationID string) *Client {
	return &Client{
		apiToken:  apiToken,
		stationID: stationID,
	}
}

type Conditions struct {
	ImperialTemperature float32
}

func (c *Client) GetCurrentConditions(ctx context.Context) (*Conditions, error) {
	// https://docs.google.com/document/d/1KGb8bTVYRsNgljnNH67AMhckY8AQT2FVwZ9urj8SWBs/edit
	url := "https://api.weather.com/v2/pws/observations/current?stationId=" + c.stationID + "&format=json&units=e&apiKey=" + c.apiToken
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("could not get current conditions: %w", err)
	}
	defer res.Body.Close()
	resData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var data struct {
		Observations []struct {
			Imperial struct {
				Temperature float32 `json:"temp"`
			} `json:"imperial"`
		} `json:"observations"`
	}
	if err := json.Unmarshal(resData, &data); err != nil {
		log.Printf("error parsing current conditions: %s", string(resData))
		return nil, fmt.Errorf("error parsing current conditions: %w", err)
	}
	if len(data.Observations) == 0 {
		log.Printf("empty observations: %s", string(resData))
		return nil, fmt.Errorf("no observations!")
	}
	return &Conditions{
		ImperialTemperature: data.Observations[0].Imperial.Temperature,
	}, nil
}
