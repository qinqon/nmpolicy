/*
 * This file is part of the nmpolicy project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2021 Red Hat, Inc.
 *
 */

package capture_test

import (
	"fmt"
	"testing"

	assert "github.com/stretchr/testify/require"

	"github.com/nmstate/nmpolicy/nmpolicy/internal/ast"
	"github.com/nmstate/nmpolicy/nmpolicy/internal/capture"
	"github.com/nmstate/nmpolicy/nmpolicy/internal/lexer"
	"github.com/nmstate/nmpolicy/nmpolicy/types"
)

func TestBasicPolicy(t *testing.T) {
	t.Run("Capture", func(t *testing.T) {
		testNoExpressions(t)
		testNoCacheAndState(t)
		testAllCapturesCached(t)
		testResolvingExpressions(t)
		testExpressionsWithPartialCache(t)
		testExpressionsWithOverCache(t)

		testLexFailure(t)
		testParseFailure(t)
		testResolveFailure(t)
	})
}

func testNoExpressions(t *testing.T) {
	t.Run("resolve with no expression", func(t *testing.T) {
		capCtrl := capture.New(lexerStub{}, parserStub{}, resolverStub{})
		resolvedCaps, err := capCtrl.Resolve(
			map[string]string{},
			map[string]types.CaptureState{"cap0": {State: []byte("some captured state")}},
			[]byte("some state"),
		)
		assert.NoError(t, err)

		assert.Nil(t, resolvedCaps)
	})
}

func testNoCacheAndState(t *testing.T) {
	t.Run("resolve with no cache and state", func(t *testing.T) {
		capCtrl := capture.New(lexerStub{}, parserStub{}, resolverStub{})
		resolvedCaps, err := capCtrl.Resolve(
			map[string]string{"cap0": "my expression"},
			map[string]types.CaptureState{},
			[]byte{},
		)
		assert.NoError(t, err)

		assert.Nil(t, resolvedCaps)
	})
}

func testAllCapturesCached(t *testing.T) {
	t.Run("resolve with all captures cached", func(t *testing.T) {
		capCache := map[string]types.CaptureState{
			"cap0": {State: []byte("some captured state")},
			"cap1": {State: []byte("another captured state")},
		}

		capCtrl := capture.New(lexerStub{}, parserStub{}, resolverStub{})
		resolvedCaps, err := capCtrl.Resolve(
			map[string]string{
				"cap0": "my expression",
				"cap1": "another expression",
			},
			capCache,
			[]byte{},
		)
		assert.NoError(t, err)

		assert.Equal(t, capCache, resolvedCaps)
	})
}

func testResolvingExpressions(t *testing.T) {
	t.Run("resolve expressions", func(t *testing.T) {
		const capID = "cap0"

		capCtrl := capture.New(lexerStub{}, parserStub{}, resolverStub{})
		resolvedCaps, err := capCtrl.Resolve(
			map[string]string{capID: "my expression"},
			map[string]types.CaptureState{},
			[]byte("some state"),
		)
		assert.NoError(t, err)

		assert.Equal(t, map[string]types.CaptureState{capID: {}}, resolvedCaps)
	})
}

func testExpressionsWithPartialCache(t *testing.T) {
	t.Run("resolve with expressions and partial cache", func(t *testing.T) {
		const capID0 = "cap0"
		const capID1 = "cap1"

		capCache := map[string]types.CaptureState{capID0: {State: []byte("some captured state")}}
		capCtrl := capture.New(lexerStub{}, parserStub{}, resolverStub{})

		resolvedCaps, err := capCtrl.Resolve(
			map[string]string{
				capID0: "my expression",
				capID1: "another expression",
			},
			capCache,
			[]byte("some state"),
		)
		assert.NoError(t, err)

		expectedCaps := capCache
		expectedCaps[capID1] = types.CaptureState{}
		assert.Equal(t, expectedCaps, resolvedCaps)
	})
}

func testExpressionsWithOverCache(t *testing.T) {
	t.Run("resolve with cache that is not included in the expressions", func(t *testing.T) {
		const capID0 = "cap0"
		const capID1 = "cap1"

		capCache := map[string]types.CaptureState{
			capID0: {State: []byte("some captured state")},
			capID1: {State: []byte("another captured state")},
		}
		capCtrl := capture.New(lexerStub{}, parserStub{}, resolverStub{})

		resolvedCaps, err := capCtrl.Resolve(
			map[string]string{
				capID0: "my expression",
			},
			capCache,
			[]byte("some state"),
		)
		assert.NoError(t, err)

		expectedCaps := map[string]types.CaptureState{
			capID0: {State: []byte("some captured state")},
		}
		assert.Equal(t, expectedCaps, resolvedCaps)
	})
}

func testLexFailure(t *testing.T) {
	t.Run("resolve fails due to lexing", func(t *testing.T) {
		capCtrl := capture.New(lexerStub{failLex: true}, parserStub{}, resolverStub{})
		_, err := capCtrl.Resolve(
			map[string]string{"cap0": "my expression"},
			map[string]types.CaptureState{},
			[]byte("some state"),
		)
		assert.Error(t, err)
	})
}

func testParseFailure(t *testing.T) {
	t.Run("resolve fails due to parsing", func(t *testing.T) {
		capCtrl := capture.New(lexerStub{}, parserStub{failParse: true}, resolverStub{})
		_, err := capCtrl.Resolve(
			map[string]string{"cap0": "my expression"},
			map[string]types.CaptureState{},
			[]byte("some state"),
		)
		assert.Error(t, err)
	})
}

func testResolveFailure(t *testing.T) {
	t.Run("resolve fails due to resolving", func(t *testing.T) {
		capCtrl := capture.New(lexerStub{}, parserStub{}, resolverStub{failResolve: true})
		_, err := capCtrl.Resolve(
			map[string]string{"cap0": "my expression"},
			map[string]types.CaptureState{},
			[]byte("some state"),
		)
		assert.Error(t, err)
	})
}

type lexerStub struct {
	failLex bool
}

func (l lexerStub) Lex(_ string) ([]lexer.Token, error) {
	if l.failLex {
		return nil, fmt.Errorf("lex failed")
	}
	return nil, nil
}

type parserStub struct {
	failParse bool
}

func (p parserStub) Parse(_ []lexer.Token) (ast.Node, error) {
	if p.failParse {
		return ast.Node{}, fmt.Errorf("parse failed")
	}
	return ast.Node{}, nil
}

type resolverStub struct {
	failResolve bool
}

func (r resolverStub) Resolve(astPool map[string]ast.Node, state []byte) (map[string]types.CaptureState, error) {
	if r.failResolve {
		return nil, fmt.Errorf("resolve failed")
	}

	capsState := map[string]types.CaptureState{}
	for id := range astPool {
		capsState[id] = types.CaptureState{}
	}

	return capsState, nil
}