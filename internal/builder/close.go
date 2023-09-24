// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

package builder

import "io"

// deferredClose is used to check the return value from Close in a defer statement.
func deferredClose(c io.Closer, err *error) {
	if cerr := c.Close(); *err == nil {
		*err = cerr
	}
}
