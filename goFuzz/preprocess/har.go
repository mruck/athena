package preprocess

import (
	"encoding/json"
	"io/ioutil"
)

type header struct {
	Name  string
	Value string
}

type cookie struct {
	Name     string
	Value    string
	Expires  string
	HTTPOnly bool
	Secure   bool
}

type param struct {
	Name  string
	Value string
}

type postData struct {
	MimeType string
	Text     string
	Params   []param
}

// Request key in har
type request struct {
	Method      string
	URL         string
	Headers     []header
	QueryString []string
	Cookies     []cookie
	PostData    postData
}

// Response key in har
type response struct {
	Status int
}

// Entry in har file
type entry struct {
	Request  request
	Response response
}

// Log in har file
type log struct {
	Entries []entry
}

// Har json representation
type Har struct {
	Log log
}

// unmarshalHar() takes in a har file and returns a Har struct
func unmarshalHar(harPath string) (*Har, error) {
	data, err := ioutil.ReadFile(harPath)
	if err != nil {
		return nil, err
	}
	har := &Har{}
	err = json.Unmarshal(data, har)
	if err != nil {
		return nil, err
	}
	return har, nil
}
