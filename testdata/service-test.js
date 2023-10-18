// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

import { check } from "k6"
import { head } from "k6/http"

export const options = {
  vus: 6,
  iterations: 12,
  thresholds: {
    checks: ['rate==1.0'],
  },
};

const archList = ["amd64", "arm64"]
const osList = ["linux", "windows", "darwin"]
const extList = ["top","dashboard","k6/x/faker", "k6/x/yaml", "k6/x/toml"]

export function setup() {
  let data = []

  const exts = extList.join(",")
    
  for(var i=0; i<archList.length; i++) {
    const arch = archList[i]
    
    for(var j=0; j<osList.length; j++) {
      const os = osList[j]
    
      data.push(`/${os}/${arch}/${exts}`)
    }
  }

  return data
}

const base = __ENV["K6X_BUILDER_SERVICE"]

export default function(data) {
  const idx = (__VU - 1) % data.length

  const res = head(`${base}${data[idx]}`)

  check(res, {'response code was 200': (res) => res.status == 200});
}
