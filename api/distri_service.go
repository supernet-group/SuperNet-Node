package api

import (
	"SuperNet-Node/pattern"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ReqModelList struct {
	Owner    string `json:"Owner"`
	Name     string `json:"Name"`
	Type1    uint8  `json:"Type1"`
	Type2    uint8  `json:"Type2"`
	OrderBy  string `json:"OrderBy"`
	Page     uint8  `json:"Page"`
	PageSize uint8  `json:"PageSize"`
}

type ResModelList struct {
	Code uint32 `json:"Code"`
	Msg  string `json:"Msg"`
	Data struct {
		List  []ResMode `json:"List"`
		Total uint32    `json:"Total"`
	} `json:"Data"`
}

type ResMode struct {
	Id         uint32    `json:"Id"`
	Owner      string    `json:"Owner"`
	Name       string    `json:"Name"`
	Framework  uint32    `json:"Framework"`
	License    uint32    `json:"License"`
	Type1      uint32    `json:"Type1"`
	Type2      uint32    `json:"Type2"`
	Tags       string    `json:"Tags"`
	CreateTime time.Time `json:"CreateTime"`
	UpdateTime time.Time `json:"UpdateTime"`
	Likes      uint32    `json:"Likes"`
	Downloads  uint32    `json:"Downloads"`
	Clicks     uint32    `json:"Clicks"`
}

func (res ResModelList) IsObsolete(fileNames []string) []string {
	m := make(map[string]bool)
	for _, mode := range res.Data.List {
		m[fmt.Sprintf("%s-%s.zip", mode.Owner, mode.Name)] = true
	}

	var obsolete []string
	for _, name := range fileNames {
		if _, ok := m[name]; !ok {
			obsolete = append(obsolete, name)
		}
	}
	return obsolete
}

func GetModelList() (ResModelList, error) {
	var resModelList ResModelList

	url := fmt.Sprintf("%s/model/list", pattern.SuperServeUrl)

	client := &http.Client{
		Timeout: time.Second * 6,
	}

	reqModelList := ReqModelList{
		Owner:    "",
		Name:     "",
		Type1:    1,
		Type2:    1,
		OrderBy:  "",
		Page:     1,
		PageSize: 10,
	}

	reqBody, err := json.Marshal(reqModelList)
	if err != nil {
		return resModelList, fmt.Errorf("> json.Marshal : %s", err.Error())
	}

	resp, err := client.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return resModelList, fmt.Errorf("> client.Post : %s", err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resModelList, fmt.Errorf("> io.ReadAll : %s", err.Error())
	}

	err = json.Unmarshal(body, &resModelList)
	if err != nil {
		return resModelList, fmt.Errorf("> json unmarshal, body: %s; err: %s", string(body), err.Error())
	}

	if resModelList.Code != 1 && resModelList.Msg != "success" {
		return resModelList, fmt.Errorf("> resModelList.Code: %d, resModelList.Msg: %s", resModelList.Code, resModelList.Msg)
	}
	return resModelList, nil
}

type Cached struct {
	Owner string `json:"Owner"`
	Name  string `json:"Name"`
}

type ReqModelCached struct {
	Owner          string `json:"Owner"`
	Uuid           string `json:"Uuid"`
	CachedModels   string `json:"CachedModels"`
	CachedDatasets string `json:"CachedDatasets"`
}

type ResModelCached struct {
	Code uint32 `json:"Code"`
	Msg  string `json:"Msg"`
}

func UpdateModelCached(ownerAddr, machineUUID string, cachedModels, cachedDatasets []Cached) (ResModelCached, error) {
	var resModelCached ResModelCached

	url := fmt.Sprintf("%s/machine/info/cached", pattern.SuperServeUrl)

	client := &http.Client{
		Timeout: time.Second * 6,
	}

	reqModelCached := ReqModelCached{
		Owner: ownerAddr,
		Uuid:  machineUUID,
	}

	dataModel, err := json.Marshal(cachedModels)
	if err != nil {
		return resModelCached, fmt.Errorf("> dataModel json.Marshal : %s", err.Error())
	}

	reqModelCached.CachedModels = string(dataModel)

	dataDataset, err := json.Marshal(cachedDatasets)
	if err != nil {
		return resModelCached, fmt.Errorf("> dataDataset json.Marshal : %s", err.Error())
	}
	reqModelCached.CachedDatasets = string(dataDataset)

	reqBody, err := json.Marshal(reqModelCached)
	if err != nil {
		return resModelCached, fmt.Errorf("> json.Marshal : %s", err.Error())
	}

	resp, err := client.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return resModelCached, fmt.Errorf("> client.Post : %s", err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resModelCached, fmt.Errorf("> io.ReadAll : %s", err.Error())
	}

	err = json.Unmarshal(body, &resModelCached)
	if err != nil {
		return resModelCached, fmt.Errorf("> json unmarshal, body: %s; err: %s", string(body), err.Error())
	}

	if resModelCached.Code != 1 && resModelCached.Msg != "success" {
		return resModelCached, fmt.Errorf("> resModelCached.Code: %d, resModelCached.Msg: %s", resModelCached.Code, resModelCached.Msg)
	}
	return resModelCached, nil
}
