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
	"github.com/nmstate/nmpolicy/nmpolicy/internal/ast"
)

func replace(inputState map[string]interface{}, pathSteps ast.VariadicOperator, replaceValue interface{}) (map[string]interface{}, error) {
	replaced, err := visitNextState(path{steps: pathSteps}, inputState, &replaceOpVisitor{replaceValue})

	if err != nil {
		return nil, replaceError("failed applying operation on the path: %w", err)
	}

	replacedMap, ok := replaced.(map[string]interface{})
	if !ok {
		return nil, replaceError("failed converting result to a map")
	}
	return replacedMap, nil
}

type replaceOpVisitor struct {
	replaceValue interface{}
}

func (r *replaceOpVisitor) visitLastMap(p path, inputMap map[string]interface{}) (interface{}, error) {
	modifiedMap := map[string]interface{}{}
	for k, v := range inputMap {
		modifiedMap[k] = v
	}

	modifiedMap[*p.currentStep.Identity] = r.replaceValue
	return modifiedMap, nil
}

func (*replaceOpVisitor) visitLastSlice(path, []interface{}) (interface{}, error) {
	return nil, nil
}

func (e *replaceOpVisitor) visitNextMap(p path, mapToVisit map[string]interface{}) (interface{}, error) {
	interfaceToVisit, ok := mapToVisit[*p.currentStep.Identity]
	if !ok {
		return nil, nil
	}

	visitResult, err := visitNextState(p, interfaceToVisit, e)
	if err != nil {
		return nil, err
	}

	replacedMap := map[string]interface{}{}
	for k, v := range mapToVisit {
		replacedMap[k] = v
	}
	replacedMap[*p.currentStep.Identity] = visitResult
	return replacedMap, nil
}

func (r *replaceOpVisitor) visitNextSlice(p path, sliceToVisit []interface{}) (interface{}, error) {
	if p.currentStep.Number != nil {
		return nil, nil
	}

	replacedSlice := make([]interface{}, len(sliceToVisit))
	for i, interfaceToVisit := range sliceToVisit {
		visitResult, err := visitNextState(p, interfaceToVisit, r)
		if err != nil {
			return nil, err
		}
		replacedSlice[i] = visitResult
	}
	return replacedSlice, nil
}
