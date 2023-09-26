// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

package builder

import (
	"archive/tar"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/szkiba/k6x/internal/resolver"

	"github.com/docker/cli/cli/connhelper"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

const xk6Image = "grafana/xk6"

func cmdline(ings resolver.Ingredients) ([]string, []string) {
	args := make([]string, 0, 2*len(ings))
	env := make([]string, 0, 1)

	env = append(env, "GOOS="+runtime.GOOS)
	args = append(args, "build")

	if ing, ok := ings.K6(); ok {
		args = append(args, ing.Tag())
		env = append(env, "K6_VERSION="+ing.Tag())
	}

	for _, ext := range ings.Extensions() {
		args = append(args, "--with", ext.Module+"@"+ext.Tag())
	}

	return args, env
}

type dockerBuilder struct {
	cli *client.Client
}

func newDockerBuilder() (*dockerBuilder, error) {
	opts := make([]client.Opt, 0, 2)

	opts = append(opts, client.WithAPIVersionNegotiation())

	host := os.Getenv(client.EnvOverrideHost) //nolint:forbidigo
	if strings.HasPrefix(host, "ssh://") {
		helper, err := connhelper.GetConnectionHelper(host)
		if err != nil {
			return nil, err
		}

		httpClient := &http.Client{Transport: &http.Transport{DialContext: helper.Dialer}}

		opts = append(
			opts,
			client.WithHTTPClient(httpClient),
			client.WithHost(helper.Host),
			client.WithDialContext(helper.Dialer),
		)
	} else {
		opts = append(opts, client.FromEnv)
	}

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, err
	}

	return &dockerBuilder{cli: cli}, nil
}

func (b *dockerBuilder) close() {
	if err := b.cli.Close(); err != nil {
		logrus.Error(err)
	}
}

func (b *dockerBuilder) pull(ctx context.Context) error {
	logrus.Debugf("Pulling %s image", xk6Image)

	reader, err := b.cli.ImagePull(ctx, xk6Image, types.ImagePullOptions{})
	if err != nil {
		return err
	}

	defer reader.Close() //nolint:errcheck

	decoder := json.NewDecoder(reader)

	for decoder.More() {
		line := make(map[string]interface{})
		if err = decoder.Decode(&line); err != nil {
			logrus.WithError(err).Error("Error while decoding docker pull output")

			break
		}

		if _, ok := line["progress"]; ok {
			continue
		}

		if status, ok := line["status"]; ok {
			delete(line, "progressDetail")
			delete(line, "status")

			e := logrus.NewEntry(logrus.StandardLogger())
			for k, v := range line {
				e = e.WithField(k, v)
			}

			e.Debug(status)
		} else {
			logrus.Debug(line)
		}
	}

	return nil
}

func (b *dockerBuilder) start(ctx context.Context, ings resolver.Ingredients) (string, error) {
	cmd, env := cmdline(ings)

	logrus.Debugf("Executing %s", strings.Join(cmd, " "))

	conf := &container.Config{
		Image: xk6Image,
		Cmd:   cmd,
		Tty:   true,
		Env:   env,
	}

	resp, err := b.cli.ContainerCreate(ctx, conf, nil, nil, nil, "")
	if err != nil {
		return "", err
	}

	logrus.Debugf("Starting container: %s", resp.ID)
	if err = b.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (b *dockerBuilder) wait(ctx context.Context, id string) error {
	statusCh, errCh := b.cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-statusCh:
	}

	return nil
}

func (b *dockerBuilder) log(ctx context.Context, id string) error {
	if !logrus.IsLevelEnabled(logrus.DebugLevel) {
		return nil
	}

	var out io.ReadCloser

	out, err := b.cli.ContainerLogs(ctx, id, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		return err
	}

	lout := logrus.StandardLogger().Writer()

	_, err = stdcopy.StdCopy(lout, lout, out)

	return err
}

func (b *dockerBuilder) copy(ctx context.Context, id string, dir string, afs afero.Fs) error {
	binary, _, err := b.cli.CopyFromContainer(ctx, id, "/xk6")
	if err != nil {
		return err
	}

	archive := tar.NewReader(binary)

	for {
		header, err := archive.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return err
		}

		if header.Typeflag == tar.TypeReg {
			fname := filepath.Join(dir, filepath.Base(header.Name))
			if runtime.GOOS == "windows" {
				fname += ".exe"
			}

			var file afero.File

			file, err = afs.OpenFile(
				fname,
				os.O_CREATE|os.O_WRONLY|os.O_TRUNC, //nolint:forbidigo
				fs.FileMode(header.Mode)|fs.ModePerm,
			)
			if err != nil {
				return err
			}

			defer deferredClose(file, &err)

			_, err = io.Copy(file, archive) //nolint:gosec
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (b *dockerBuilder) Build(
	ctx context.Context,
	ings resolver.Ingredients,
	dir string,
	afs afero.Fs,
) error {
	defer b.close()

	return b.build(ctx, ings, dir, afs)
}

func (b *dockerBuilder) build(
	ctx context.Context,
	ings resolver.Ingredients,
	dir string,
	afs afero.Fs,
) error {
	logrus.Debug("Building new k6 binary (docker)")

	if err := afs.MkdirAll(dir, 0o750); err != nil {
		return err
	}

	if err := b.pull(ctx); err != nil {
		return err
	}

	id, err := b.start(ctx, ings)
	if err != nil {
		return err
	}

	defer func() {
		logrus.Debugf("Removing container: %s", id)

		rerr := b.cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{})
		if rerr != nil && err == nil {
			err = rerr
		}
	}()

	if err = b.wait(ctx, id); err != nil {
		return err
	}

	if err = b.log(ctx, id); err != nil {
		return err
	}

	if err = b.copy(ctx, id, dir, afs); err != nil {
		return err
	}

	return err
}
