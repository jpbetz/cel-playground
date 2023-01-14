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
	"encoding/json"
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

type EvalResponse struct {
	Result string `json:"result"`
}

func handleEval(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type")

	switch r.Method {
	case "OPTIONS":
		w.WriteHeader(204)
		return
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
		json, err := json.Marshal(&EvalResponse{Result: fmt.Sprintf("%+v", result)})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(json)
	default:
		http.Error(w, "Supported methods: POST", http.StatusMethodNotAllowed)
	}
}
