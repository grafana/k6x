"use k6 >= v0.50";
"use k6 with k6/x/faker >= 0.3";

import http from "k6/http";
import faker from "k6/x/faker";

export default function () {
  const user = faker.person.name();

  const resp = http.post("https://httpbin.test.k6.io/post", JSON.stringify(user));
}
