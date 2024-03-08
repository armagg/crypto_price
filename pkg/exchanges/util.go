package exchanges

import (
	"encoding/json"
	"sync"

)

type MapMutex struct {
    Map map[interface{}]interface{}
    Mutex sync.Mutex    
}

func (mm *MapMutex) SetAsync(key interface{}, val interface{}){
    mm.Mutex.Lock()
    defer mm.Mutex.Unlock()
    mm.Map[key] = val
}

func (mm *MapMutex) Get(key interface{}) (interface{}, error) {
    
    return mm.Map[key], nil
}

type PriceTicker struct {
    Symbol string  `json:"symbol"`
    Price  float64 `json:"price"`
}

func UnmarshalJSON(data []byte, v interface{}) error {
    return json.Unmarshal(data, v)
}

