package boc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const bocDataLink = "https://www.banqueducanada.ca/valet/observations/group/bond_yields_all/json"

type BOCInterests interface {
	GetObservationForDate(date string) (*Observations, error)
	GroupDetail() GroupDetail
	Terms() Terms
	SeriesDetail() SeriesDetail
}

type bocInterests struct {
	data         *BOCData
	observations map[string]*Observations
	url          string
}

// NewBOCInterests provides an interface to get the interests data from Bank of Canada
func NewBOCInterests() (BOCInterests, error) {
	boc := new(bocInterests)
	boc.url = bocDataLink
	if err := boc.fetchData(); err != nil {
		return nil, fmt.Errorf("error fetching data: %w", err)
	}
	boc.setObservationsMap()
	return boc, nil
}

// GroupDetail implements BOCInterests
func (b *bocInterests) GroupDetail() GroupDetail {
	return b.data.GroupDetail
}

// Terms implements BOCInterests
func (b *bocInterests) Terms() Terms {
	return b.data.Terms
}

// SeriesDetail implements BOCInterests
func (b *bocInterests) SeriesDetail() SeriesDetail {
	return b.data.SeriesDetail
}

func (b *bocInterests) setObservationsMap() {
	m := make(map[string]*Observations)
	for _, obs := range b.data.Observations {
		obs := obs
		m[obs.D] = &obs
	}
	b.observations = m
}

// GetObservationForDate implements BOCInterests
func (b *bocInterests) GetObservationForDate(date string) (*Observations, error) {
	date, err := FormatDate(date)

	if err != nil {
		return nil, fmt.Errorf("invalid date format: %s", date)
	}
	if b.observations[date] == nil {
		return nil, fmt.Errorf("no data for this date: %s", date)
	}
	return b.observations[date], nil
}

// FormatDate formats a date string according to what is expected for boc's data
func FormatDate(date string) (string, error) {
	date = strings.TrimSpace(date)
	separator := ""
	if strings.Contains(date, "-") {
		separator = "-"
	} else if strings.Contains(date, "\\") {
		separator = "\\"
	} else if strings.Contains(date, "/") {
		separator = "/"
	}

	parts := strings.Split(date, separator)
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	digits := make([]int, 0)
	year, month, day, yearInd := 0, 0, 0, 0
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid number of parts in date: %v", len(parts))
	}

	// validate digits
	for _, p := range parts {
		nb, err := strconv.Atoi(p)
		if err != nil {
			return "", fmt.Errorf("part should be digit: %s", p)
		}
		if nb < 0 {
			return "", fmt.Errorf("part cannot be negative: %s", p)
		}
		digits = append(digits, nb)

	}

	// find year
	for i, p := range parts {
		if len(p) == 4 {
			year, _ = strconv.Atoi(p)
			yearInd = i
			for j, d := range digits {
				if year == d {
					parts = append(parts[:i], parts[i+1:]...)
					digits = append(digits[:j], digits[j+1:]...)
				}
			}
			break
		}
	}

	// find day/month
	if digits[0] > 12 {
		day = digits[0]
		month = digits[1]
	} else if digits[1] > 12 {
		day = digits[1]
		month = digits[0]
	} else if yearInd == 0 {
		month = digits[0]
		day = digits[1]
	} else if yearInd == 2 {
		if digits[0] > digits[1] {
			month = digits[1]
			day = digits[0]
		} else {
			month = digits[0]
			day = digits[1]
		}
	}

	if year == 0 || month == 0 || day == 0 {
		return "", fmt.Errorf("invalid format: %s", date)
	}
	if month > 12 || month < 0 {
		return "", fmt.Errorf("invalid month: %d", month)
	}

	if day > 31 || day < 0 {
		return "", fmt.Errorf("invalid day: %d", day)
	}
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day), nil
}

func (b *bocInterests) fetchData() error {
	resp, err := http.Get(bocDataLink)
	if err != nil {
		return fmt.Errorf("error fetching data: %w", err)
	}
	respData, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error reading body data")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid Response code: %v\n\nResp data: %v", resp.StatusCode, string(respData))
	}
	jsonData := new(BOCData)
	if err = json.Unmarshal(respData, jsonData); err != nil {
		return fmt.Errorf("failed to parse json data")
	}
	b.data = jsonData
	return nil
}

type BOCData struct {
	GroupDetail  GroupDetail    `json:"groupDetail"`
	Terms        Terms          `json:"terms"`
	SeriesDetail SeriesDetail   `json:"seriesDetail"`
	Observations []Observations `json:"observations"`
}

type Observations struct {
	D                 string `json:"d"`
	YieldRRB          Val    `json:"BD.CDN.RRB.DQ.YLD,omitempty"`
	Average5To10Year  Val    `json:"CDN.AVG.5YTO10Y.AVG,omitempty"`
	Yield3Year        Val    `json:"BD.CDN.3YR.DQ.YLD,omitempty"`
	Yield10Year       Val    `json:"BD.CDN.10YR.DQ.YLD"`
	Average3To5Year   Val    `json:"CDN.AVG.3YTO5Y.AVG,omitempty"`
	Yield2Year        Val    `json:"BD.CDN.2YR.DQ.YLD,omitempty"`
	Yield7Year        Val    `json:"BD.CDN.7YR.DQ.YLD,omitempty"`
	Average1To3Year   Val    `json:"CDN.AVG.1YTO3Y.AVG,omitempty"`
	AverageOver10Year Val    `json:"CDN.AVG.OVER.10.AVG,omitempty"`
	Yield5Year        Val    `json:"BD.CDN.5YR.DQ.YLD,omitempty"`
	YieldLong         Val    `json:"BD.CDN.LONG.DQ.YLD,omitempty"`
}
type SeriesDetail struct {
	Average1To3Year   Detail `json:"CDN.AVG.1YTO3Y.AVG"`
	Average3To5Year   Detail `json:"CDN.AVG.3YTO5Y.AVG"`
	Average5To10Year  Detail `json:"CDN.AVG.5YTO10Y.AVG"`
	AverageOver10Year Detail `json:"CDN.AVG.OVER.10.AVG"`
	Yield2Year        Detail `json:"BD.CDN.2YR.DQ.YLD"`
	Yield3Year        Detail `json:"BD.CDN.3YR.DQ.YLD"`
	Yield5Year        Detail `json:"BD.CDN.5YR.DQ.YLD"`
	Yield7Year        Detail `json:"BD.CDN.7YR.DQ.YLD"`
	Yield10Year       Detail `json:"BD.CDN.10YR.DQ.YLD"`
	YieldLong         Detail `json:"BD.CDN.LONG.DQ.YLD"`
	YieldRRB          Detail `json:"BD.CDN.RRB.DQ.YLD"`
}
type GroupDetail struct {
	Label       string `json:"label"`
	Description string `json:"description"`
	Link        string `json:"link"`
}
type Detail struct {
	Label       string    `json:"label"`
	Description string    `json:"description"`
	Dimension   Dimension `json:"dimension"`
}

type Dimension struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}
type Terms struct {
	URL string `json:"url"`
}
type Val struct {
	V string `json:"v"`
}

func hasSameData(bocAll, boc bocInterests) error {
	mAll := make(map[string]*Observations)
	for _, obs := range bocAll.observations {
		mAll[obs.D] = obs
	}
	m := make(map[string]*Observations)
	for _, obs := range boc.observations {
		m[obs.D] = obs
	}

	if len(mAll) != len(m) {
		return fmt.Errorf("data doesn't have the same count: %v vs %v", len(mAll), len(m))
	}

	for date, obsAll := range mAll {
		if err := isSameObs(obsAll, m[date]); err != nil {
			return fmt.Errorf("data not the same for date: %v \n %v \n vs \n %v\n, error: %w", date, obsAll, m[date], err)
		}
	}

	return nil
}

func isSameObs(obsAll, obs *Observations) error {
	valsAll := []Val{
		obsAll.Average1To3Year,
		obsAll.Average3To5Year,
		obsAll.Average5To10Year,
		obsAll.AverageOver10Year,
		obsAll.Yield2Year,
		obsAll.Yield3Year,
		obsAll.Yield5Year,
		obsAll.Yield7Year,
		obsAll.Yield10Year,
		obsAll.YieldLong,
		obsAll.YieldRRB,
	}
	vals := []Val{
		obs.Average1To3Year,
		obs.Average3To5Year,
		obs.Average5To10Year,
		obs.AverageOver10Year,
		obs.Yield2Year,
		obs.Yield3Year,
		obs.Yield5Year,
		obs.Yield7Year,
		obs.Yield10Year,
		obs.YieldLong,
		obs.YieldRRB,
	}
	for i, v := range vals {
		if v.V != "" && v.V != valsAll[i].V {
			return fmt.Errorf("vals are not the same, all: %v vs %v", valsAll[i].V, v.V)
		}
	}
	return nil
}
