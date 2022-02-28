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

	"github.com/nmstate/nmpolicy/nmpolicy/internal/ast"
)

func walk(inputState map[string]interface{}, pathSteps ast.VariadicOperator) (interface{}, error) {
	visitResult, err := visitNextState(path{steps: pathSteps}, inputState, &walkOpVisitor{})
	if err != nil {
		return nil, fmt.Errorf("failed walking path: %w", err)
	}

	return visitResult, nil
}

type walkOpVisitor struct{}

func (*walkOpVisitor) visitLastMap(p path, mapToAccess map[string]interface{}) (interface{}, error) {
	v, ok := mapToAccess[*p.currentStep.Identity]
	if !ok {
		return nil, pathError(p.currentStep, "step not found at map state '%+v'", mapToAccess)
	}
	return v, nil
}

func (*walkOpVisitor) visitLastSlice(p path, sliceToAccess []interface{}) (interface{}, error) {
	if len(sliceToAccess) <= *p.currentStep.Number {
		return nil, pathError(p.currentStep, "step not found at slice state '%+v'", sliceToAccess)
	}
	return sliceToAccess[*p.currentStep.Number], nil
}

func (w *walkOpVisitor) visitNextSlice(p path, sliceToVisit []interface{}) (interface{}, error) {
	if p.currentStep.Number == nil {
		return nil, pathError(p.peekNextStep(), "unexpected non numeric step for slice state '%+v'", sliceToVisit)
	}
	interfaceToVisit, err := w.visitLastSlice(p, sliceToVisit)
	if err != nil {
		return nil, wrapWithPathError(p.currentStep, err)
	}
	return visitNextState(p, interfaceToVisit, w)
}

func (w *walkOpVisitor) visitNextMap(p path, mapToVisit map[string]interface{}) (interface{}, error) {
	interfaceToVisit, err := w.visitLastMap(p, mapToVisit)
	if err != nil {
		return nil, wrapWithPathError(p.currentStep, err)
	}
	return visitNextState(p, interfaceToVisit, w)
}
