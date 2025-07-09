package gomodproxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/pentops/log.go/log"
)

type Info struct {
	Version string
	Time    time.Time
}

var ErrNotImplemented = fmt.Errorf("not implemented")

type VersionNotFoundError string

func (vnfe VersionNotFoundError) Error() string {
	return fmt.Sprintf("version not found: %s", string(vnfe))
}

func (vnfe VersionNotFoundError) HTTPError(r *http.Request, w http.ResponseWriter) {
	http.Error(w, vnfe.Error(), http.StatusNotFound)
}

type httpError interface {
	HTTPError(r *http.Request, w http.ResponseWriter)
}

type StatusError struct {
	Code int
	Err  error
}

func (e StatusError) Error() string {
	return e.Err.Error()
}

func (e StatusError) HTTPError(r *http.Request, w http.ResponseWriter) {
	http.Error(w, e.Err.Error(), e.Code)
}

type RawResponse struct {
	Code        int
	Body        []byte
	Reader      io.ReadCloser
	ContentType string
}

func (rr RawResponse) HTTPResponse(r *http.Request, w http.ResponseWriter) {
	if rr.ContentType != "" {
		w.Header().Set("Content-Type", rr.ContentType)
	}
	w.WriteHeader(rr.Code)

	if rr.Reader != nil {
		defer rr.Reader.Close()
		io.Copy(w, rr.Reader) // nolint: errcheck
		return
	}

	w.Write(rr.Body) // nolint: errcheck
}

type ModProvider interface {
	GoModLatest(ctx context.Context, packageName string) (*Info, error)
	GoModList(ctx context.Context, packageName string) ([]string, error)

	GoModInfo(ctx context.Context, packageName, version string) (*Info, error)
	GoModMod(ctx context.Context, packageName, version string) ([]byte, error)
	GoModZip(ctx context.Context, packageName, version string) (io.ReadCloser, error)
}

type Command int

const (
	LatestCommand Command = iota
	ListCommand
	InfoCommand
	ModCommand
	ZipCommand
)

var commandMap = map[string]Command{
	".info": InfoCommand,
	".mod":  ModCommand,
	".zip":  ZipCommand,
}

type ParsedRequest struct {
	PackageName string
	Version     string
	Command     Command
}

func ParseRequestPath(requestPath string) (ParsedRequest, error) {
	// go mod has a strange path structure which isn't compatible with any
	// popular routing library.
	// $package/@latest
	// $package/@v/list
	// $package/@v/$version.$ext
	// Where $package is a go import path, including '/' characters.

	request := ParsedRequest{}

	parts := strings.Split(requestPath, "/")
	if len(parts) < 2 || parts[0] != "" {
		return request, StatusError{Code: http.StatusNotFound, Err: fmt.Errorf("invalid path")}
	}
	parts = parts[1:]
	if parts[0] == "gopkg" {
		parts = parts[1:]
	}

	moduleParts := make([]string, 0, len(parts)-1)
	found := false
	for idx, part := range parts {
		if part == "" {
			return request, StatusError{Code: http.StatusNotFound, Err: fmt.Errorf("invalid path")}
		}
		if part[0] == '@' {
			parts = parts[idx:]
			found = true
			break
		}
		moduleParts = append(moduleParts, part)
	}
	if !found {
		return request, StatusError{Code: http.StatusNotFound, Err: fmt.Errorf("no /@ found in path")}
	}

	request.PackageName = strings.Join(moduleParts, "/")

	if parts[0] == "@latest" {
		request.Command = LatestCommand
		return request, nil
	}

	if parts[0] != "@v" {
		return request, fmt.Errorf("expecting @v")
	}

	if parts[1] == "list" {
		request.Command = ListCommand
		return request, nil
	}

	fileType := path.Ext(parts[1])
	versionName := parts[1][:len(parts[1])-len(fileType)]
	request.Version = versionName

	var ok bool
	request.Command, ok = commandMap[fileType]
	if !ok {
		return request, fmt.Errorf("invalid file type '%s'", fileType)
	}

	return request, nil

}

func Handler(mods ModProvider) http.Handler {

	sendJSON := func(w http.ResponseWriter, code int, data any) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		dataJSON, err := json.Marshal(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(code)
		w.Write(dataJSON) // nolint: errcheck
	}

	sendError := func(w http.ResponseWriter, r *http.Request, err error) {
		if he, ok := err.(httpError); ok {
			he.HTTPError(r, w)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK\n")) // nolint: errcheck
			return
		}
		ctx := r.Context()
		ctx = log.WithFields(ctx, map[string]any{
			"method": r.Method,
			"host":   r.Host,
			"path":   r.URL.Path,
		})
		log.Info(ctx, "request")

		parsed, err := ParseRequestPath(r.URL.Path)
		if err != nil {
			log.WithError(ctx, err).Error("http error")
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		switch parsed.Command {
		case LatestCommand:
			info, err := mods.GoModLatest(ctx, parsed.PackageName)
			if err != nil {
				if errors.Is(err, ErrNotImplemented) {
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}
				log.WithError(ctx, err).Error("http error")
				sendError(w, r, err)
				return
			}
			log.Info(ctx, "OK")
			sendJSON(w, http.StatusOK, info)
			return

		case ListCommand:
			versions, err := mods.GoModList(ctx, parsed.PackageName)
			if err != nil {
				log.WithError(ctx, err).Error("http error")
				sendError(w, r, err)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(strings.Join(versions, "\n"))) // nolint: errcheck
			return

		case InfoCommand:
			info, err := mods.GoModInfo(ctx, parsed.PackageName, parsed.Version)
			if err != nil {
				log.WithError(ctx, err).Error("http error")
				sendError(w, r, err)
				return
			}
			log.Info(ctx, "OK")
			sendJSON(w, http.StatusOK, info)
			return

		case ModCommand:
			mod, err := mods.GoModMod(ctx, parsed.PackageName, parsed.Version)
			if err != nil {
				log.WithError(ctx, err).Error("http error")
				sendError(w, r, err)
				return
			}

			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write(mod) // nolint: errcheck
			return

		case ZipCommand:
			zip, err := mods.GoModZip(ctx, parsed.PackageName, parsed.Version)
			if err != nil {
				log.WithError(ctx, err).Error("http error")
				sendError(w, r, err)
				return
			}

			w.Header().Set("Content-Type", "application/zip")
			w.WriteHeader(http.StatusOK)
			io.Copy(w, zip) // nolint: errcheck
			return

		default:
			sendError(w, r, fmt.Errorf("unknown command %v", parsed.Command))
		}

	})
}

func Serve(ctx context.Context, port int, mods ModProvider) error {

	mux := http.NewServeMux()

	mux.Handle("/gopkg/", http.StripPrefix("/gopkg", Handler(mods)))

	ss := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return ss.ListenAndServe()
}
