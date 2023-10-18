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
