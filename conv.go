package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"log"

	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	ep "github.com/wrossmorrow/envoy-extproc-sdk-go"
)

type convRequestProcessor struct {
	opts *ep.ProcessingOptions
}

func (s *convRequestProcessor) GetName() string {
	return "xml-json-conv"
}

func (s *convRequestProcessor) GetOptions() *ep.ProcessingOptions {
	return s.opts
}

func (s *convRequestProcessor) ProcessRequestHeaders(ctx *ep.RequestContext, headers ep.AllHeaders) error {
	return ctx.ContinueRequest()
}

const (
	kXML  = "xml"
	kJSON = "json"
)

// detectFormat attempts to determine whether the data is JSON or XML.
func detectFormat(data []byte) (string, error) {
	if json.Valid(data) {
		return kJSON, nil
	}

	var xmlObj any
	if err := xml.Unmarshal(data, &xmlObj); err == nil {
		return kXML, nil
	}

	return "", errors.New("unknown format")
}

func (s *convRequestProcessor) ProcessRequestBody(ctx *ep.RequestContext, body []byte) error {
	cancel := func(code int32) error {
		return ctx.CancelRequest(code, map[string]ep.HeaderValue{}, typev3.StatusCode_name[code])
	}

	from, err := detectFormat(body)
	if err != nil {
		log.Println(err)
		return cancel(400)
	}

	var data []byte

	switch from {
	case kJSON:
		data, err = jsonToXML(body)
	case kXML:
		data, err = xmlToJSON(body)
	default:
		panic("never happen")
	}

	if err != nil {
		log.Printf("convert data format is failed: %v", err)
		return cancel(400)
	}

	return ctx.CancelRequest(200, map[string]ep.HeaderValue{}, string(data))
}

func (s *convRequestProcessor) ProcessRequestTrailers(ctx *ep.RequestContext, trailers ep.AllHeaders) error {
	return ctx.ContinueRequest()
}

func (s *convRequestProcessor) ProcessResponseHeaders(ctx *ep.RequestContext, headers ep.AllHeaders) error {
	return ctx.ContinueRequest()
}

func (s *convRequestProcessor) ProcessResponseBody(ctx *ep.RequestContext, body []byte) error {
	return ctx.ContinueRequest()
}

func (s *convRequestProcessor) ProcessResponseTrailers(ctx *ep.RequestContext, trailers ep.AllHeaders) error {
	return ctx.ContinueRequest()
}

func (s *convRequestProcessor) Init(opts *ep.ProcessingOptions, nonFlagArgs []string) error {
	s.opts = opts
	return nil
}

func (s *convRequestProcessor) Finish() {}

// xmlToJSON converts XML data to JSON.
func xmlToJSON(data []byte) ([]byte, error) {
	var m map[string]any
	err := xml.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}
	return json.Marshal(m)
}

// jsonToXML converts JSON data to XML.
func jsonToXML(data []byte) ([]byte, error) {
	var m map[string]any
	err := json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}
	return xml.Marshal(m)
}
