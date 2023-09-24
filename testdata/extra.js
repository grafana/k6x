// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

"use k6 >= v0.46"
"use k6 with k6/x/faker >= 0.2"

import faker from "k6/x/faker"

import parts from "./extra/parts.js"

export default function () {
  let user = faker.person()
}
