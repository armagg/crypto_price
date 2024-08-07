package config

import (
	"log"
	"os"

	yaml "gopkg.in/yaml.v3"
)

func GetConfigs()(map[string]string){
	f, err := os.ReadFile("pkg/config/env.yml")

    if err != nil {
        log.Fatal(err)
    }

    var data map[string]string

    err = yaml.Unmarshal(f, &data)

    if err != nil {
        log.Fatal(err)
    }

    return data
}