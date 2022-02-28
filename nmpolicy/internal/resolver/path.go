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

type captureEntryNameAndSteps struct {
	captureEntryName string
	steps            ast.VariadicOperator
}

type path struct {
	steps            []ast.Node
	currentStepIndex int
	currentStep      *ast.Node
}

type stateVisitor interface {
	visitLastMap(path, map[string]interface{}) (interface{}, error)
	visitLastSlice(path, []interface{}) (interface{}, error)
	visitNextMap(path, map[string]interface{}) (interface{}, error)
	visitNextSlice(path, []interface{}) (interface{}, error)
}

func visitNextState(p path, inputState interface{}, v stateVisitor) (interface{}, error) {
	p.nextStep()
	originalMap, isMap := inputState.(map[string]interface{})
	if isMap {
		if p.hasMoreSteps() {
			if p.currentStep.Identity == nil {
				return nil, pathError(p.currentStep, "unexpected non identity step for map state '%+v'", originalMap)
			}
			return v.visitNextMap(p, originalMap)
		}
		return v.visitLastMap(p, originalMap)
	}

	originalSlice, isSlice := inputState.([]interface{})
	if isSlice {
		if p.hasMoreSteps() || p.currentStep.Number == nil {
			if p.currentStep.Number == nil {
				p.backStep()
			}
			return v.visitNextSlice(p, originalSlice)
		}
		return v.visitLastSlice(p, originalSlice)
	}
	return nil, pathError(p.currentStep, "invalid type %T for identity step '%v'", inputState, *p.currentStep)
}

func (p *path) nextStep() {
	if p.currentStep == nil {
		p.currentStepIndex = 0
	} else if p.hasMoreSteps() {
		p.currentStepIndex++
	}
	p.currentStep = &p.steps[p.currentStepIndex]
}

func (p *path) backStep() {
	if p.currentStep == nil {
		p.currentStepIndex = 0
	} else if p.currentStepIndex > 0 {
		p.currentStepIndex--
	}
	p.currentStep = &p.steps[p.currentStepIndex]
}

func (p *path) hasMoreSteps() bool {
	return p.currentStepIndex+1 < len(p.steps)
}

func (p *path) peekNextStep() *ast.Node {
	if !p.hasMoreSteps() {
		return &p.steps[p.currentStepIndex]
	}
	return &p.steps[p.currentStepIndex+1]
}

func (p captureEntryNameAndSteps) walkState(stateToWalk map[string]interface{}) (interface{}, error) {
	return walk(stateToWalk, p.steps)
}
