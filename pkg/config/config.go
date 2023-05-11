package config

import (
	"fmt"
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

    fmt.Println(data)
    return data
}