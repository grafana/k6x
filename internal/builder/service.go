// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

package builder

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/szkiba/k6x/internal/dependency"
)

const serviceTimeout = 60 * time.Second

var (
	errService         = errors.New("service error")
	errServiceEndpoint = errors.New("missing build service endpoint")
)

type serviceBuilder struct {
	client *http.Client
}

func newServiceBuilder(ctx context.Context) (Builder, bool, error) {
	if buildServiceDisabled(ctx) {
		return nil, false, nil
	}

	client := &http.Client{Timeout: serviceTimeout}

	return &serviceBuilder{client: client}, true, nil
}

func (b *serviceBuilder) Engine() Engine {
	return Service
}

func (b *serviceBuilder) Build(
	ctx context.Context,
	platform *Platform,
	mods dependency.Modules,
	out io.Writer,
) error {
	if platform == nil {
		platform = RuntimePlatform()
	}

	return b.build(ctx, platform, mods, out)
}

func builderService() string {
	return os.Getenv("K6X_BUILDER_SERVICE") //nolint:forbidigo
}

func buildServiceDisabled(ctx context.Context) bool {
	loc := builderService()
	if len(loc) == 0 {
		return true
	}

	u, err := url.Parse(loc)
	if err != nil {
		return false
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var r net.Resolver

	txts, err := r.LookupTXT(ctx, u.Hostname())
	if err != nil {
		return false
	}

	for _, txt := range txts {
		if txt == "disabled=true" {
			return true
		}
	}

	return false
}

func (b *serviceBuilder) build(
	ctx context.Context,
	platform *Platform,
	mods dependency.Modules,
	out io.Writer,
) error {
	logrus.Debug("Building new k6 binary (service)")

	service := builderService()
	if len(service) == 0 {
		return errServiceEndpoint
	}

	path := "/" + platform.String() + "/" + mods.ToArtifacts().String()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, service+path, nil)
	if err != nil {
		return err
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %s", errService, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %s", errService, resp.Status)
	}

	defer resp.Body.Close() //nolint:errcheck

	if _, err = io.Copy(out, resp.Body); err != nil {
		return err
	}

	return nil
}
