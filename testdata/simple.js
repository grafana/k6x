// SPDX-FileCopyrightText: 2023 Iván SZKIBA
//
// SPDX-License-Identifier: AGPL-3.0-only

import faker from "k6/x/faker"

import parts from "./simple/parts.js"

export default function () {
  let user = faker.person()
}
