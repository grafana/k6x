"use k6 >= v0.46"
"use k6 with k6/x/faker >= 0.2"
import http from "k6/http"

import faker from "k6/x/faker" // not included in official k6 binary

export default function () {
  let user = faker.person()

  let resp = http.post("https://httpbin.test.k6.io/post", JSON.stringify(user))
}
