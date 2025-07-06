// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/lrstanley/go-ytdlp"
	sloghttp "github.com/samber/slog-http"
)

// Example curl call:
//   $ curl -sS \
//         -H "Content-Type: application/json" \
//         --data @example-request-body.json \
//         http://localhost:8080/download

var downloadsPath = "/tmp/ytdlp-downloads"

// RequestBody is an example of how you might structure a request.
type RequestBody struct {
	Env   map[string]string `json:"env,omitempty"`
	Flags ytdlp.FlagConfig  `json:"flags"`
	Args  []string          `json:"args"`
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	logger.Info("creating downloads path", "path", downloadsPath)
	if err := os.MkdirAll(downloadsPath, 0o750); err != nil {
		logger.Error("failed to create downloads path", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/download", postDownload)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      sloghttp.New(logger)(mux),
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
	}

	logger.Info("starting server", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}

func jsonError(w http.ResponseWriter, r *http.Request, code int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	data := map[string]any{
		"error": err.Error(),
		"code":  code,
	}

	if perr, ok := ytdlp.IsMultipleJSONParsingFlagsError(err); ok {
		data["errors"] = perr.Errors
	}

	if perr, ok := ytdlp.IsJSONParsingFlagError(err); ok {
		data["error"] = perr.Err.Error()
		data["id"] = perr.ID
		data["json_path"] = perr.JSONPath
		data["flag"] = perr.Flag
	}

	if err := json.NewEncoder(w).Encode(&data); err != nil {
		slog.ErrorContext(r.Context(), "failed to encode error", "error", err)
		return
	}
}

func postDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Unmarshal JSON into RequestBody.
	var body RequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 2. Validate flags using [FlagConfig.Validate], or using the JSON schema (noted above).
	if err := body.Flags.Validate(); err != nil {
		jsonError(w, r, http.StatusBadRequest, err)
		return
	}

	// 3. Validate user provides at least one positional argument.
	if len(body.Args) == 0 {
		http.Error(w, "at least one positional argument is required", http.StatusBadRequest)
		return
	}

	// 4. Use the values to construct and run a yt-dlp command, by calling `Command.SetFlagConfig`,
	//    `Command.SetEnvVar`, and then `Command.Run`.
	cmd := ytdlp.New().
		SetFlagConfig(&body.Flags)

	if body.Flags.Filesystem.Output != nil {
		cmd.Output(filepath.Join(
			downloadsPath,
			*body.Flags.Filesystem.Output,
		))
	} else {
		cmd.Output(filepath.Join(
			downloadsPath,
			"%(extractor)s - %(title)s.%(ext)s",
		))
	}

	if body.Env != nil {
		for k, v := range body.Env {
			// You may want to allow-list certain env vars that people can provide, for security reasons.
			cmd.SetEnvVar(k, v)
		}
	}

	// print current run flags.
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	err := enc.Encode(cmd.GetFlagConfig())
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to encode flags", "error", err)
		return
	}

	// 5. Run the command. Ideally, this handler would return a response immediately, with another endpoint
	//    to get the status of the download, and the associated results, as it may take longer to download
	//    than the user is willing to wait, and/or the timeout value would allow.
	result, err := cmd.Run(context.Background(), body.Args...)
	if err != nil {
		jsonError(w, r, http.StatusUnprocessableEntity, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		slog.ErrorContext(r.Context(), "failed to encode result", "error", err)
		return
	}
}
