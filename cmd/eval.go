/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2" //"sigs.k8s.io/yaml"
	"k8s.io/apiserver/pkg/cel/library"
)

var evalCmd = &cobra.Command{
	Use:   "eval",
	Short: "Evaluate a CEL expression.",
	RunE: func(cmd *cobra.Command, args []string) error {
		vars, err := parseVariablesFlag(evalOpts.variables)
		if err != nil {
			return err
		}
		return evalAndPrint(evalOpts.expression, vars)
	},
}

type evalOptions struct {
	expression string
	variables  string
}

var evalOpts = &evalOptions{}

func init() {
	RootCmd.AddCommand(evalCmd)
	evalCmd.Flags().StringVar(&evalOpts.expression, "expr", "", "CEL expression to evaluate")
	evalCmd.Flags().StringVar(&evalOpts.variables, "variables", "", "Comma seperated list of <variable-name>=<YAML filename> pairs.")
}

func baseEnv() (*cel.Env, error) {
	var opts []cel.EnvOption
	opts = append(opts, cel.HomogeneousAggregateLiterals())
	// Validate function declarations once during base env initialization,
	// so they don't need to be evaluated each time a CEL rule is compiled.
	// This is a relatively expensive operation.
	opts = append(opts, cel.EagerlyValidateDeclarations(true), cel.DefaultUTCTimeZone(true))
	opts = append(opts, library.ExtensionLibs...)

	return cel.NewEnv(opts...)
}

func evalAndPrint(expression string, vars map[string]any) error {
	result, err := eval(expression, vars)
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", result)
	return nil
}

func eval(expression string, vars map[string]any) (any, error) {
	var envOpts []cel.EnvOption
	baseEnv, err := baseEnv()
	if err != nil {
		return nil, err
	}
	activation := map[string]any{}
	for variableName, data := range vars {
		envOpts = append(envOpts, cel.Variable(variableName, cel.DynType))
		activation[variableName] = data
	}
	env, err := baseEnv.Extend(envOpts...)
	if err != nil {
		return nil, err
	}
	ast, issues := env.Compile(expression)
	if issues != nil {
		return nil, issues.Err()
	}
	program, err := baseEnv.Program(ast, cel.OptimizeRegex(library.ExtensionLibRegexOptimizations...))
	if err != nil {
		return nil, err
	}
	result, _, err := program.Eval(activation)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func parseVariablesFlag(flag string) (map[string]any, error) {
	if len(flag) == 0 {
		return nil, nil
	}
	result := map[string]any{}
	for _, e := range strings.Split(flag, ",") {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("expected <variable-name>=<filename> but got: %s", e)
		}
		varName := parts[0]
		filename := parts[1]
		parsedYAML, err := parseYAMLFile(varName, filename)
		if err != nil {
			return nil, err
		}
		result[parsedYAML.variableName] = parsedYAML.data
	}
	return result, nil
}

type parsedYAML struct {
	variableName string
	data         map[string]any
}

func parseYAMLFile(varName, filename string) (*parsedYAML, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading input file %s: %v", filename, err)
	}
	parsed := map[string]any{}
	err = yaml.Unmarshal(data, parsed)
	if err != nil {
		return nil, err
	}
	return &parsedYAML{variableName: varName, data: parsed}, nil
}
