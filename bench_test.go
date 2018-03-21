// Copyright © 2018 Douglas Chimento <dchimento@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lambdazap

import (
	"io/ioutil"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// A Syncer is a spy for the Sync portion of zapcore.WriteSyncer.
type Syncer struct {
	err    error
	called bool
}

// SetError sets the error that the Sync method will return.
func (s *Syncer) SetError(err error) {
	s.err = err
}

// Sync records that it was called, then returns the user-supplied error (if
// any).
func (s *Syncer) Sync() error {
	s.called = true
	return s.err
}

// Called reports whether the Sync method was called.
func (s *Syncer) Called() bool {
	return s.called
}

// A Discarder sends all writes to ioutil.Discard.
type Discarder struct{ Syncer }

// Write implements io.Writer.
func (d *Discarder) Write(b []byte) (int, error) {
	return ioutil.Discard.Write(b)
}

func BenchmarkStatic(b *testing.B) {
	setStatics()
	defer func() {
		reset()
	}()
	lc := New(ProcessNonContextFields(false)).With(FunctionName, FunctionVersion, AwsRequestID)
	blogger := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionConfig().EncoderConfig),
			&Discarder{},
			zap.DebugLevel,
		))
	blogger.With(lc.NonContextValues()...)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			blogger.Info("test")
		}
	})
}

func BenchmarkContextValues(b *testing.B) {
	setStatics()
	defer func() {
		reset()
	}()
	lbc, cf := getContext()
	defer cf()
	lc := New(ProcessNonContextFields(true)).WithBasic()
	blogger := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionConfig().EncoderConfig),
			&Discarder{},
			zap.DebugLevel,
		))
	blogger.With(lc.NonContextValues()...)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			blogger.Info("test", lc.ContextValues(lbc)...)
		}
	})
}

func BenchmarkAll(b *testing.B) {
	setStatics()
	defer func() {
		reset()
	}()
	lbc, cf := getContext()
	defer cf()
	lc := New(ProcessNonContextFields(true)).WithAll()
	blogger := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionConfig().EncoderConfig),
			&Discarder{},
			zap.DebugLevel,
		))
	blogger.With(lc.NonContextValues()...)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			blogger.Info("test", lc.ContextValues(lbc)...)
		}
	})
}
