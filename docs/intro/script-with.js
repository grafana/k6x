"use k6 >= v0.47"
"use k6 with k6/x/dotenv >= 0.1.3"
"use k6 with k6/x/yaml >= 0.1.4"
"use k6 with k6/x/ansible-vault >= 0.1.3"
"use k6 with k6/x/faker >= 0.2"
import { post } from "k6/http"
import "k6/x/dotenv"
import YAML from "k6/x/yaml"
import { decrypt } from "k6/x/ansible-vault"
import faker from "k6/x/faker"

const secrets = open("secrets.vault")

export function setup() {
  return { secrets:YAML.parse(decrypt(secrets, __ENV["SECRET"])) }
}

export default function({ secrets }) {
  const user = faker.person()

  const resp = post("https://httpbin.test.k6.io/post", JSON.stringify(user), {
    headers: { Authorization: `Bearer ${secrets.prod.apikey}` }
  })
}
