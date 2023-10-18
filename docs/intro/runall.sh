#!/bin/bash

# SPDX-FileCopyrightText: 2023 Iván SZKIBA
#
# SPDX-License-Identifier: AGPL-3.0-only

for exp in *.exp; do
  rm -f k6 $HOME/.cache/k6x/bin/k6
  ./$exp
  rm -f "$(basename -s .exp $exp).cast"
done