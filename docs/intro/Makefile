# SPDX-FileCopyrightText: 2023 Iván SZKIBA
#
# SPDX-License-Identifier: AGPL-3.0-only

all: script.svg script-with.svg top.svg readme.svg get.svg get-lazy.svg cloud.svg filter.svg build.svg

.SILENT:

%.cast: %.exp
	rm -f k6 $$HOME/.cache/k6x/bin/k6
	./$<

%.svg: %.cast
	svg-term --in $< --out $@ --width `head -1 $<|jq .width` --height `head -1 $<|jq .height`
