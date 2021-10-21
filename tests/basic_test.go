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

package tests

import (
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"

	"github.com/nmstate/nmpolicy/nmpolicy"
)

func TestBasicPolicy(t *testing.T) {
	t.Run("Basic policy", func(t *testing.T) {
		testEmptyPolicy(t)
		testPolicyWithOnlyDesiredState(t)
	})
}

func testEmptyPolicy(t *testing.T) {
	t.Run("is empty", func(t *testing.T) {
		s, err := nmpolicy.GenerateState(nmpolicy.PolicySpec{}, nil, nmpolicy.NoCache())

		assert.NoError(t, err)

		expectedEmptyState := nmpolicy.GeneratedState{MetaInfo: nmpolicy.MetaInfo{Version: "0"}}
		assert.NotEqual(t, time.Time{}, s.MetaInfo.TimeStamp)
		assert.Equal(t, expectedEmptyState, resetTimeStamp(s))
	})
}

func testPolicyWithOnlyDesiredState(t *testing.T) {
	// When a basic input with only the desired state is provided,
	// the policy just passes it as is to the output with no modifications.
	t.Run("with only desired state", func(t *testing.T) {
		stateData := []byte(`this is not a legal yaml format!`)
		policySpec := nmpolicy.PolicySpec{
			DesiredState: stateData,
		}

		s, err := nmpolicy.GenerateState(policySpec, nil, nmpolicy.NoCache())

		assert.NoError(t, err)
		expectedState := nmpolicy.GeneratedState{
			DesiredState: stateData,
			MetaInfo:     nmpolicy.MetaInfo{Version: "0"},
		}
		assert.Equal(t, expectedState, resetTimeStamp(s))
	})
}

func resetTimeStamp(s nmpolicy.GeneratedState) nmpolicy.GeneratedState {
	s.MetaInfo.TimeStamp = time.Time{}
	return s
}