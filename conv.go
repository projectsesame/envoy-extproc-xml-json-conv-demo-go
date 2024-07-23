package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"log"

	mxj "github.com/clbanning/mxj/v2"
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

	if isValidXML(data) {
		return kXML, nil
	}

	return "", errors.New("unknown format")
}

func isValidXML(input []byte) bool {
	decoder := xml.NewDecoder(bytes.NewReader(input))
	for {
		err := decoder.Decode(new(any))
		if err != nil {
			return err == io.EOF
		}
	}
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

// jsonToXML converts a JSON byte array to XML byte array.
func jsonToXML(data []byte) ([]byte, error) {
	mxj.XMLEscapeChars(true)
	m, err := mxj.NewMapJsonReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	err = m.XmlIndentWriter(buf, "", "\t")
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// xmlToJSON converts XML data to JSON.
func xmlToJSON(data []byte) ([]byte, error) {
	mxj.XMLEscapeChars(true)
	m, err := mxj.NewMapXmlReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	err = m.JsonIndentWriter(buf, "", "\t")
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
