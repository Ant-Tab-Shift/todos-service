package models

import "net/http"

type Endpoint struct {
	Pattern string
	Func    http.HandlerFunc
}
