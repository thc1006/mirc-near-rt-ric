// Copyright 2017 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"

	"github.com/emicklei/go-restful/v3"
)

// JSONPool provides reusable JSON encoders and decoders to reduce allocations
type JSONPool struct {
	encoderPool sync.Pool
	decoderPool sync.Pool
	bufferPool  sync.Pool
}

// GlobalJSONPool is the global instance for reuse across handlers
var GlobalJSONPool = NewJSONPool()

// NewJSONPool creates a new JSON pool
func NewJSONPool() *JSONPool {
	return &JSONPool{
		encoderPool: sync.Pool{
			New: func() interface{} {
				buf := &bytes.Buffer{}
				return json.NewEncoder(buf)
			},
		},
		decoderPool: sync.Pool{
			New: func() interface{} {
				return json.NewDecoder(nil)
			},
		},
		bufferPool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}
}

// EncodeJSON encodes an object to JSON using pooled resources
func (jp *JSONPool) EncodeJSON(v interface{}) ([]byte, error) {
	buf := jp.bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer jp.bufferPool.Put(buf)

	encoder := jp.encoderPool.Get().(*json.Encoder)
	encoder.SetOutput(buf)
	defer jp.encoderPool.Put(encoder)

	if err := encoder.Encode(v); err != nil {
		return nil, err
	}

	// Copy buffer content since we're returning it after putting buffer back to pool
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result, nil
}

// DecodeJSON decodes JSON from reader using pooled resources
func (jp *JSONPool) DecodeJSON(r io.Reader, v interface{}) error {
	decoder := jp.decoderPool.Get().(*json.Decoder)
	decoder.Reset(r)
	defer jp.decoderPool.Put(decoder)

	return decoder.Decode(v)
}

// WriteJSONResponse writes a JSON response using pooled resources
func (jp *JSONPool) WriteJSONResponse(response *restful.Response, v interface{}) error {
	data, err := jp.EncodeJSON(v)
	if err != nil {
		return err
	}

	response.Header().Set("Content-Type", "application/json")
	_, err = response.Write(data)
	return err
}

// ReadJSONRequest reads a JSON request using pooled resources
func (jp *JSONPool) ReadJSONRequest(request *restful.Request, v interface{}) error {
	return jp.DecodeJSON(request.Request.Body, v)
}

// OptimizedJSONMiddleware provides optimized JSON handling middleware
func OptimizedJSONMiddleware(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	// Add pooled JSON utilities to request context
	req.SetAttribute("jsonPool", GlobalJSONPool)
	chain.ProcessFilter(req, resp)
}

// Helper functions for backward compatibility

// WriteEntity writes an entity using optimized JSON encoding
func WriteEntityOptimized(response *restful.Response, value interface{}) error {
	return GlobalJSONPool.WriteJSONResponse(response, value)
}

// ReadEntityOptimized reads an entity using optimized JSON decoding
func ReadEntityOptimized(request *restful.Request, value interface{}) error {
	return GlobalJSONPool.ReadJSONRequest(request, value)
}

// StreamingJSONEncoder provides streaming JSON encoding for large datasets
type StreamingJSONEncoder struct {
	encoder *json.Encoder
	writer  io.Writer
}

// NewStreamingJSONEncoder creates a new streaming JSON encoder
func NewStreamingJSONEncoder(w io.Writer) *StreamingJSONEncoder {
	return &StreamingJSONEncoder{
		encoder: json.NewEncoder(w),
		writer:  w,
	}
}

// WriteArrayStart writes the start of a JSON array
func (sje *StreamingJSONEncoder) WriteArrayStart() error {
	_, err := sje.writer.Write([]byte("["))
	return err
}

// WriteArrayEnd writes the end of a JSON array
func (sje *StreamingJSONEncoder) WriteArrayEnd() error {
	_, err := sje.writer.Write([]byte("]"))
	return err
}

// WriteComma writes a comma separator
func (sje *StreamingJSONEncoder) WriteComma() error {
	_, err := sje.writer.Write([]byte(","))
	return err
}

// WriteItem writes a single item to the array
func (sje *StreamingJSONEncoder) WriteItem(item interface{}) error {
	return sje.encoder.Encode(item)
}

// WriteItems writes multiple items as a JSON array with streaming
func (sje *StreamingJSONEncoder) WriteItems(items []interface{}) error {
	if err := sje.WriteArrayStart(); err != nil {
		return err
	}

	for i, item := range items {
		if i > 0 {
			if err := sje.WriteComma(); err != nil {
				return err
			}
		}
		if err := sje.WriteItem(item); err != nil {
			return err
		}
	}

	return sje.WriteArrayEnd()
}

// JSONSizeOptimizer provides methods to optimize JSON output size
type JSONSizeOptimizer struct {
	omitEmpty bool
}

// NewJSONSizeOptimizer creates a new JSON size optimizer
func NewJSONSizeOptimizer(omitEmpty bool) *JSONSizeOptimizer {
	return &JSONSizeOptimizer{omitEmpty: omitEmpty}
}

// OptimizeStruct optimizes struct fields for JSON serialization
func (jso *JSONSizeOptimizer) OptimizeStruct(v interface{}) interface{} {
	// This would typically use reflection to remove empty fields
	// For now, return as-is, but in production this could be implemented
	// to dynamically remove nil/empty fields before serialization
	return v
}

// CompactJSON removes unnecessary whitespace from JSON
func CompactJSON(data []byte) []byte {
	buf := &bytes.Buffer{}
	if err := json.Compact(buf, data); err != nil {
		return data // Return original if compaction fails
	}
	return buf.Bytes()
}