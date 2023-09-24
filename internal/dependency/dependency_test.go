// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

package dependency_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/szkiba/k6x/internal/dependency"
)

func TestNew(t *testing.T) {
	t.Parallel()

	ctor := dependency.New

	var dep *dependency.Dependency
	var err error

	dep, err = ctor("", "")

	assert.NotNil(t, dep)
	assert.NoError(t, err)

	dep, err = ctor("foo", "> v0.1.0")

	assert.NotNil(t, dep)
	assert.NoError(t, err)

	assert.Equal(t, "foo", dep.Name)
	assert.Equal(t, ">v0.1.0", dep.Constraints.String())
}

func TestNew_error(t *testing.T) {
	t.Parallel()

	ctor := dependency.New

	var dep *dependency.Dependency
	var err error

	dep, err = ctor("foo", ">bar")

	assert.Nil(t, dep)
	assert.ErrorIs(t, err, dependency.ErrInvalidConstraints)
}
