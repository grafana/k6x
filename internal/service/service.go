// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package service contains build service implementation.
//
//nolint:revive
package service

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/szkiba/k6x/internal/builder"
	"github.com/szkiba/k6x/internal/resolver"
)

type service struct {
	resolver resolver.Resolver
	builder  builder.Builder
}

func New(r resolver.Resolver, b builder.Builder) http.Handler {
	svc := new(service)

	svc.resolver = r
	svc.builder = b

	return svc
}

func (svc *service) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		res.WriteHeader(http.StatusMethodNotAllowed)

		return
	}

	params, err := parseParams(req.URL.Path)
	if err != nil {
		if errors.Is(err, errInvalidParameters) {
			svc.looseServeHTTP(res, req)

			return
		}

		http.Error(res, err.Error(), http.StatusBadRequest)

		return
	}

	log := logrus.WithField("params", params.String())

	canonical := params.String()

	if req.URL.Path != canonical {
		log.WithField("from", req.URL.Path).WithField("action", "redirect").Info()

		setCacheControl(res)
		http.Redirect(res, req, canonical, http.StatusMovedPermanently)

		return
	}

	if req.Header.Get("If-None-Match") == params.ETag() {
		log.WithField("action", "skip").Info()

		res.WriteHeader(http.StatusNotModified)

		return
	}

	mods, err := svc.resolver.Resolve(req.Context(), params.ToDependencies())
	if err != nil {
		log.WithError(err).Error("resolve error")

		http.Error(res, err.Error(), http.StatusBadRequest)

		return
	}

	var buff bytes.Buffer

	log.WithField("action", "build").Info()

	if err := svc.builder.Build(req.Context(), params.Platform, mods, &buff); err != nil {
		log.WithError(err).Error("build error")

		http.Error(res, err.Error(), http.StatusPreconditionFailed)

		return
	}

	setHeaders(res, params, buff.Len())

	_, _ = io.Copy(res, &buff)
}

func (svc *service) looseServeHTTP(res http.ResponseWriter, req *http.Request) {
	params, err := looseParseParams(req.Context(), req.URL.Path, svc.resolver)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)

		return
	}

	canonical := params.String()

	logrus.WithField("params", canonical).
		WithField("from", req.URL.Path).
		WithField("action", "resolve").
		Info()

	res.Header().Set("Cache-Control", "no-cache,no-store")
	http.Redirect(res, req, canonical, http.StatusTemporaryRedirect)
}

func setCacheControl(res http.ResponseWriter) {
	res.Header().
		Set("Cache-Control",
			fmt.Sprintf(
				"public, max-age=%d, immutable, stale-while-revalidate=%d, stale-if-error=%d",
				maxAge,
				maxAgeStale,
				maxAgeStale,
			),
		)
}

func setHeaders(res http.ResponseWriter, params *Params, contentLength int) {
	setCacheControl(res)
	res.Header().Set("Content-Length", strconv.Itoa(contentLength))
	res.Header().Set("ETag", params.ETag())
	res.Header().Set("Content-Type", "application/octet-stream")

	ext := ""
	if params.OS == "windows" {
		ext = ".exe"
	}

	res.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="k6%s"`, ext))
}

const (
	maxAge      = 60 * 60 // ~ 1 hour
	maxAgeStale = 60 * 10 // ~ 10 mins
	maxAgeError = 60 * 5  // ~ 5 mins

	// maxAge   = 60 * 60 * 24 * 30 * 3 // ~ 3 month
	// maxAgeStale = 60 * 60 * 24 * 7      // ~ 1 weeks
	// maxAgeError = 60 * 5  // ~ 5 mins
)
