/*
 * Copyright 2021 NMPolicy Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *
 *	  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package resolver

import (
	"fmt"
	"reflect"

	"github.com/nmstate/nmpolicy/nmpolicy/internal/ast"
)

var (
	filterLookupMap = map[string]bool{
		"interfaces": true,
		"routes":     true,
		"running":    true,
		"config":     true,
	}
)

func filter(inputState map[string]interface{}, pathSteps ast.VariadicOperator, expectedValue interface{}) (map[string]interface{}, error) {
	filtered, err := visitNextState(path{steps: pathSteps}, inputState, &eqFilterVisitor{expectedValue})

	if err != nil {
		return nil, fmt.Errorf("failed applying operation on the path: %w", err)
	}

	if filtered == nil {
		return nil, nil
	}

	filteredMap, ok := filtered.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed converting filtering result to a map")
	}
	return filteredMap, nil
}

type eqFilterVisitor struct {
	expectedValue interface{}
}

func (e *eqFilterVisitor) visitLastMap(p path, mapToFilter map[string]interface{}) (interface{}, error) {
	obtainedValue, ok := mapToFilter[*p.currentStep.Identity]
	if !ok {
		return nil, nil
	}
	if reflect.TypeOf(obtainedValue) != reflect.TypeOf(e.expectedValue) {
		return nil, pathError(p.currentStep, `type missmatch: the value in the path doesn't match the value to filter. `+
			`"%T" != "%T" -> %+v != %+v`, obtainedValue, e.expectedValue, obtainedValue, e.expectedValue)
	}
	if obtainedValue == e.expectedValue {
		return mapToFilter, nil
	}
	return nil, nil
}

func (e *eqFilterVisitor) visitLastSlice(path, []interface{}) (interface{}, error) {
	return nil, nil
}

func (e *eqFilterVisitor) visitNextMap(p path,
	mapToVisit map[string]interface{}) (interface{}, error) {
	interfaceToVisit, ok := mapToVisit[*p.currentStep.Identity]
	if !ok {
		return nil, nil
	}

	visitResult, err := visitNextState(p, interfaceToVisit, e)
	if err != nil {
		return nil, err
	}
	if visitResult == nil {
		return nil, nil
	}
	filteredMap := map[string]interface{}{}
	if !shouldFilter(p.currentStep) {
		for k, v := range mapToVisit {
			filteredMap[k] = v
		}
	}
	filteredMap[*p.currentStep.Identity] = visitResult
	return filteredMap, nil
}

func (e *eqFilterVisitor) visitNextSlice(p path, sliceToVisit []interface{}) (interface{}, error) {
	if p.currentStep.Number != nil {
		return nil, nil
	}

	filteredSlice := []interface{}{}
	hasVisitResult := false
	for _, interfaceToVisit := range sliceToVisit {
		visitResult, err := visitNextState(p, interfaceToVisit, e)
		if err != nil {
			return nil, err
		}
		if visitResult != nil {
			hasVisitResult = true
			filteredSlice = append(filteredSlice, visitResult)
		} else if !shouldFilter(p.currentStep) {
			filteredSlice = append(filteredSlice, interfaceToVisit)
		}
	}

	if !hasVisitResult {
		return nil, nil
	}
	return filteredSlice, nil
}

func shouldFilter(currentStep *ast.Node) bool {
	if currentStep == nil {
		return true
	}
	if currentStep.Identity == nil {
		return false
	}
	_, ok := filterLookupMap[*currentStep.Identity]
	return ok
}
