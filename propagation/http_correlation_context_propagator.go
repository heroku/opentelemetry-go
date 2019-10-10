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

package propagation

import (
	"context"
	"net/url"
	"strings"

	"go.opentelemetry.io/api/core"
	"go.opentelemetry.io/api/key"
	apipropagation "go.opentelemetry.io/api/propagation"
)

type httpCorrrelationContextKeyType int

const (
	CorrelationContextHeader = "Correlation-Context"

	correlationDataKey httpCorrrelationContextKeyType = iota
)

type HTTPCorrelationContextPropagator struct {
	KV []core.KeyValue
}

var _ apipropagation.TextFormatPropagator = HTTPCorrelationContextPropagator{}

func WithCorrelationContextInfo(ctx context.Context, kv ...core.KeyValue) context.Context {
	return context.WithValue(ctx, correlationDataKey, kv)
}

func getCorrelationsContextInfo(ctx context.Context) []core.KeyValue {
	return ctx.Value(correlationDataKey).([]core.KeyValue)
}

func (p HTTPCorrelationContextPropagator) Inject(ctx context.Context, supplier apipropagation.Supplier) {
	var headerValueBuilder strings.Builder
	firstIter := true
	for _, kv := range getCorrelationsContextInfo(ctx) {
		if !firstIter {
			headerValueBuilder.WriteRune(',')
		}
		firstIter = false
		headerValueBuilder.WriteString(url.QueryEscape(strings.TrimSpace(kv.Key.Name)))
		headerValueBuilder.WriteRune('=')
		headerValueBuilder.WriteString(url.QueryEscape(strings.TrimSpace(kv.Value.Emit())))
	}
	supplier.Set(CorrelationContextHeader, headerValueBuilder.String())
}

func (hp HTTPCorrelationContextPropagator) Extract(ctx context.Context, supplier apipropagation.Supplier) core.SpanContext {
	correlationContext := supplier.Get(CorrelationContextHeader)
	if correlationContext == "" {
		return core.SpanContext{}
	}

	contextValues := strings.Split(correlationContext, ",")
	var sc core.SpanContext
	sc.KV = make([]core.KeyValue, 0, len(contextValues))
	for _, contextValue := range contextValues {
		valueAndProps := strings.Split(contextValue, ";")
		if len(valueAndProps) < 1 {
			continue
		}
		nameValue := strings.Split(valueAndProps[0], "=")
		if len(nameValue) < 2 {
			continue
		}
		name, err := url.QueryUnescape(nameValue[0])
		if err != nil {
			continue
		}
		trimmedName := strings.TrimSpace(name)
		value, err := url.QueryUnescape(nameValue[1])
		if err != nil {
			continue
		}
		trimmedValue := strings.TrimSpace(value)

		// TODO (skaris): properties defiend https://w3c.github.io/correlation-context/, are currently
		// just put as part of the value.
		var trimmedValueWithProps strings.Builder
		trimmedValueWithProps.WriteString(trimmedValue)
		for _, prop := range valueAndProps[1:] {
			trimmedValueWithProps.WriteRune(';')
			trimmedValueWithProps.WriteString(prop)
		}

		sc.KV = append(sc.KV, key.New(trimmedName).String(trimmedValueWithProps.String()))
	}
	return sc
}

func (hp HTTPCorrelationContextPropagator) GetAllKeys() []string {
	return []string{CorrelationContextHeader}
}
