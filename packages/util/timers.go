// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"time"
)

type TimerParam struct {
	name         string
	defaultValue time.Duration
	value        *time.Duration
}

func NewTimerParam(n string, dv time.Duration) *TimerParam {
	return &TimerParam{
		name:         n,
		defaultValue: dv,
		value:        nil,
	}
}

func (tpT *TimerParam) GetName() string {
	return tpT.name
}

func (tpT *TimerParam) GetDefault() time.Duration {
	return tpT.defaultValue
}

func (tpT *TimerParam) getValue() time.Duration {
	if tpT.value == nil {
		return tpT.GetDefault()
	}
	return *(tpT.value)
}

func (tpT *TimerParam) setValue(v time.Duration) {
	tpT.value = &v
}

func (tpT *TimerParam) clearValue() {
	tpT.value = nil
}

type TimerParams map[string]*TimerParam

func NewTimerParams(params ...*TimerParam) TimerParams {
	result := make(map[string]*TimerParam)
	for _, param := range params {
		result[param.GetName()] = param
	}
	return result
}

func (tpT TimerParams) Get(name string) time.Duration {
	return tpT[name].getValue()
}

func (tpT TimerParams) Set(name string, value time.Duration) {
	tpT[name].setValue(value)
}

func (tpT TimerParams) With(name string, value time.Duration) TimerParams {
	tpT.Set(name, value)
	return tpT
}

func (tpT TimerParams) Clear(name string) {
	tpT[name].clearValue()
}
