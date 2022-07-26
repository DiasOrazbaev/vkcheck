package vk

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

var CountryCodeTo = make(map[string]string, 237)

type Phone struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func Generate() {
	/*
		var phones []Phone

		j = strings.TrimSpace(j)

		err := json.Unmarshal([]byte(j), &phones)
		if err != nil {
			return
		}

		for _, phone := range phones {
			CountryCodeTo[strings.ReplaceAll(phone.Code, " ", "")] = phone.Name
		}

		k, err := json.Marshal(CountryCodeTo)
		if err != nil {
			return
		}
		f, err := os.Create("phones.json")

		_, _ = fmt.Fprint(f, string(k))*/

	b, err := ioutil.ReadFile("phones.json")
	if err != nil {
		log.Fatalln(err)
	}
	err = json.Unmarshal(b, &CountryCodeTo)
	if err != nil {
		log.Fatalln(err)
	}
}
