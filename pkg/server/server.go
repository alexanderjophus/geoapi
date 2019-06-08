package server

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"
)

var postcodesUrl = "https://api.postcodes.io/postcodes/"

var postcodesClient = http.Client{
	Timeout: 5 * time.Second,
}

type PostcodeResponse struct {
	Status  int     `json:"status"`
	LongLat LongLat `json:"result"`
}

type LongLat struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

func Run() {
	log.Println("service running")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		postcode := r.URL.Query().Get("postcode")
		if postcode == "" {
			return
		}
		longLat, err := getLongLat(postcode)
		if err != nil {
			w.Write([]byte(err.Error()))
		}

		providers, err := calculateProviders(longLat)
		if err != nil {
			w.Write([]byte(err.Error()))
		}
		ret, err := json.Marshal(providers)
		if err != nil {
			w.Write([]byte(err.Error()))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(200)
		w.Write(ret)
	})

	http.ListenAndServe(":80", nil)
	return
}

func getLongLat(postcode string) (*LongLat, error) {
	url := fmt.Sprintf(postcodesUrl + strings.ReplaceAll(postcode, " ", ""))
	resp, err := postcodesClient.Get(url)
	if err != nil {
		return nil, err
	}
	var result PostcodeResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result.LongLat, nil
}

type Providers struct {
	Name       string  `json:"Name"`
	Address    string  `json:"Address"`
	Postcode   string  `json:"Postal"`
	Categories string  `json:"Categories"`
	Miles      int     `json:"Miles"`
	Type       string  `json:"Type"`
	LongLat    LongLat `json:"longlat"`
	Distance   float64 `json:"distance"`
}

func calculateProviders(source *LongLat) ([]Providers, error) {
	providers, err := getProviders()
	if err != nil {
		return nil, err
	}
	for i, provider := range providers {
		longLat, err := getLongLat(provider.Postcode)
		if err != nil {
			return nil, err
		}
		providers[i].LongLat = *longLat
		providers[i].Distance = getDistance(*source, *longLat)
	}
	return providers, nil
}

func getProviders() ([]Providers, error) {
	var providers []Providers
	r := strings.NewReader(`[  {    "Name": "Avon and Wiltshire Mental Health Partnership NHS Trust",    "Address": "Bath NHS House, Newbridge Hill, Bath ",    "Postal": "BA1 3QE",    "Categories": "Healthcare",    "Miles": 100,    "Type": "Operational"  },  {    "Name": "Royal United Hospitals Bath NHS Foundation Trust",    "Address": "Combe Park, Bath",    "Postal": "BA1 3NG",    "Categories": "Healthcare, counselling, psychotherapy",    "Miles": 10,    "Type": "Operational"  },  {    "Name": "Weldmar Hospice",    "Address": "Herringston Road, Dorchester",    "Postal": "DT1 2SL",    "Categories": "Bereavement, end of Life, counselling, psychotherapy, hospice",    "Miles": 34,    "Type": "Operational"  },  {    "Name": "Sturminster Newton Medical Centre",    "Address": "Old Market Hill, Sturminster Newton",    "Postal": "DT10 1QU",    "Categories": "Repite Care",    "Miles": 64,    "Type": "Operational"  },  {    "Name": "We the Curious",    "Address": "ne Millennium Square, Anchor Rd, Bristol ",    "Postal": "BS1 5DB",    "Categories": "Special Days,Complimentary Services, Hairdressing",    "Miles": 37,    "Type": "Operational"  },  {    "Name": "Bristol Zoo",    "Address": "Bristol",    "Postal": "BS8 3HA",    "Categories": "Respite Care, Special Days",    "Miles": 800,    "Type": "Operational"  },  {    "Name": "Oakham Treasures",    "Address": "Oakham farm",    "Postal": "BS20 7SP",    "Categories": "Respite Care, Special Days",    "Miles": 244,    "Type": "Operational"  },  {    "Name": "Batch Golf Club",    "Address": "Sham Castle, Golf Course Rd, Bath ",    "Postal": "BA2 6JG",    "Categories": "Sport, Special Days, Recovery Support",    "Miles": 23,    "Type": "Operational"  },  {    "Name": "Headscarves By Ciara",    "Address": "7 Bridgelea cottages , Newtownards , North Down , United Kingdom",    "Postal": "BT23 7TQ",    "Categories": "Accessories, Support, Clothing, Complimentary therapy",    "Miles": 50,    "Type": "Operational"  },  {    "Name": "Duffus Cancer Foundation",    "Address": "Duffus street",    "Postal": "",    "Categories": "Support, Bereavement, Carer, Charity",    "Miles": null,    "Type": "On-line"  },  {    "Name": "Ebisu Health Limited",    "Address": "Support, Bereavement, Counselling, Healthcare",    "Postal": "",    "Categories": "",    "Miles": null,    "Type": ""  } ]`)
	err := json.NewDecoder(r).Decode(&providers)
	return providers, err
}

func getDistance(x, y LongLat) float64 {

	radLat1 := math.Pi * x.Latitude / 180
	radLat2 := math.Pi * y.Latitude / 180

	theta := x.Longitude - y.Longitude
	radTheta := math.Pi * theta / 180

	dist := math.Sin(radLat1)*math.Sin(radLat2) + math.Cos(radLat1)*math.Cos(radLat2)*math.Cos(radTheta)

	if dist > 1 {
		dist = 1
	}

	dist = math.Acos(dist)
	dist = dist * 180 / math.Pi
	dist = dist * 60 * 1.1515

	// lat := math.Pow((x.Latitude - y.Latitude), 2)
	// long := math.Pow((x.Longitude - y.Longitude), 2)
	// return math.Sqrt(lat + long)
	return dist
}
