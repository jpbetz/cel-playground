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
	"io"
	"net/http"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serves a CEL playground.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return serve(serveOpts.port)
	},
}

type serveOptions struct {
	port int32
	// TODO: add TLS flags
}

var serveOpts = &serveOptions{}

func init() {
	RootCmd.AddCommand(serveCmd)
	serveCmd.Flags().Int32Var(&serveOpts.port, "port", 8080, "HTTP port")
}

func serve(port int32) error {
	http.HandleFunc("/eval", handleEval)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		return err
	}
	return nil
}

type EvalRequest struct {
	Expression string         `json:"expression"`
	Variables  map[string]any `json:"variables"`
}

func handleEval(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	switch r.Method {
	case "POST":
		contentType := r.Header.Get("Content-type")
		switch contentType {
		case "application/json", "application/yaml":
			// Follow k8s content-type conventions
		default:
			http.Error(w, "Supported Content-Type values: application/json, application/yaml", http.StatusUnsupportedMediaType)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		request := &EvalRequest{}
		err = yaml.Unmarshal(body, request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := eval(request.Expression, request.Variables)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, "%v", result)
	default:
		http.Error(w, "Supported methods: POST", http.StatusMethodNotAllowed)
	}
}
