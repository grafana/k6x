"use k6 >= v0.50";
"use k6 with k6/x/faker >= 0.3";

import faker from "k6/x/faker";

import parts from "./simple/parts.js";

export default function () {
  let user = faker.person();
}
