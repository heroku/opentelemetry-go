// Copyright 2019, OpenTelemetry Authors
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

package stdout

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"sync"

	"go.opentelemetry.io/sdk/export"
	"go.opentelemetry.io/sdk/trace"
)

// Options are the options to be used when initializing a stdout export.
type Options struct {
	// PrettyPrint will pretty the json representation of the span,
	// making it print "pretty". Default is false.
	PrettyPrint bool
}

// Exporter is an implementation of trace.Exporter that writes spans to stdout.
type Exporter struct {
	once         sync.Once
	pretty       bool
	outputWriter io.Writer
}

func NewExporter(o Options) (*Exporter, error) {
	return &Exporter{
		pretty:       o.PrettyPrint,
		outputWriter: os.Stdout,
	}, nil
}

// RegisterSimpleSpanProcessor registers e as SimpleSpanProcessor.
func (e *Exporter) RegisterSimpleSpanProcessor() {
	e.once.Do(func() {
		ssp := trace.NewSimpleSpanProcessor(e)
		trace.RegisterSpanProcessor(ssp)
	})
}

// ExportSpan writes a SpanData in json format to stdout.
func (e *Exporter) ExportSpan(ctx context.Context, data *export.SpanData) {
	var jsonSpan []byte
	var err error
	if e.pretty {
		jsonSpan, err = json.MarshalIndent(data, "", "\t")
	} else {
		jsonSpan, err = json.Marshal(data)
	}
	if err != nil {
		// ignore writer failures for now
		_, _ = e.outputWriter.Write([]byte("Error converting spanData to json: " + err.Error()))
		return
	}
	// ignore writer failures for now
	_, _ = e.outputWriter.Write(append(jsonSpan, byte('\n')))
}
