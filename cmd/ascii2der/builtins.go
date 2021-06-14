// Copyright 2015 The DER ASCII Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"errors"
	"fmt"
)

// NOTE: If adding a builtin, remember to document it in language.txt!

var builtins = map[string]func(*scanner, [][]byte) ([]byte, error){
	// define(var, val) sets var = val in the scanner's variable table.
	// Variables may be redefined. Expands to the empty string.
	"define": func(scanner *scanner, args [][]byte) ([]byte, error) {
		if len(args) != 2 {
			return nil, errors.New("expected two arguments to define()")
		}
		scanner.vars[string(args[0])] = args[1]
		return nil, nil
	},

	// var(var) expands to whatever var is set to in the scanner's variable table.
	// Error if var is not defined.
	//
	// var(var, default) behaves similarly, except expands to default if var is
	// not defined.
	"var": func(scanner *scanner, args [][]byte) ([]byte, error) {
		switch len(args) {
		case 1:
			val, ok := scanner.vars[string(args[0])]
			if !ok {
				return nil, fmt.Errorf("var() with undefined name: %q", string(args[0]))
			}
			return val, nil
		case 2:
			val, ok := scanner.vars[string(args[0])]
			if !ok {
				return args[1], nil
			}
			return val, nil
		default:
			return nil, errors.New("expected one or two arguments to var()")
		}
	},
}

func executeBuiltin(scanner *scanner, name string, args [][]byte) ([]byte, error) {
	builtin, ok := builtins[name]
	if !ok {
		return nil, fmt.Errorf("unrecognized builtin %q", name)
	}

	return builtin(scanner, args)
}
