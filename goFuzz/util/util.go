package util

import (
	"encoding/json"
	"fmt"
	"log"
)

// PrettyPrint a struct
func PrettyPrint(data interface{}) {
	jsonified, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		err = fmt.Errorf("failed to pretty print json: %v", err)
		log.Fatal(err)
	}
	fmt.Println(string(jsonified))
}
