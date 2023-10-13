---
author: Iván SZKIBA
marp: true
theme: uncover
---

> I just want to run my tests!

# k6x

**Run k6 with extensions as easy as possible**

*Iván Szkiba*

https://github.com/szkiba/k6x

---

# Agenda

- Use Cases
- Under The Hood
- Proposals

---

# Use Cases

What is k6x?

---

> I just want to run my test!
 
```js
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
```
---

> I just want to run my test!

```bash
k6x run script.js
```

![script.js](script.svg)

---

> I want to run my test, but with...

```js
"use k6 >= v0.47"
"use k6 with k6/x/dotenv >= 0.1"
"use k6 with k6/x/yaml >= 0.1"
"use k6 with k6/x/ansible-vault >= 0.1"
"use k6 with k6/x/faker >= 0.1"
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
```

---

> I want to run my test, but with...

```bash
k6x run script-with.js
```

![script-with.js](script-with.svg)

---

> I want to run my test, but with output extension

```bash
k6x run -o top -d 5s script.js
```

![top.js](top.svg)

---

> I want to use k6 extensions in the cloud

```bash
docker run --rm -it -e K6X_BUILDER_SERVICE=$K6X_BUILDER_SERVICE -v $PWD:/home/k6x szkiba/k6x run script.js
```

![cloud](cloud.svg)

---

> I want to control which extensions are allowed

```bash
k6x run --filter "[?contains(tiers,'Official')]" script.js
```

![filter](filter.svg)

---

> I need a k6 binary with extensions

```bash
curl -OJ $K6X_BUILDER_SERVICE/linux/amd64/k6@v0.47.0,dashboard@v0.6.0,k6/x/faker@v0.2.2,top@v0.1.1
chmod +x k6
./k6 version
```

![get](get.svg)

---

> I need a k6 binary with extensions, but I'm lazy

```bash
curl -OJL $K6X_BUILDER_SERVICE/linux/amd64/dashboard,top,k6/x/faker
chmod +x k6
./k6 version
```

![get-lazy](get-lazy.svg)

---

> I want to build k6 binary with extensions

```bash
k6x build --with dashboard --with k6/x/faker --with top --with k6/x/yaml
./k6 version
```

![build](build.svg)

---

## So what k6x is?

- k6 launcher with extensions used by the test
    - available as CLI and docker image
    - without any additional configuration
    - without installing additional tooling
      - but supports the installed toolings
        - go, Docker Engine (even remote)
- k6 native/docker builder
- k6 builder/download service

---

# Under The Hood

---

## Design Considerations

- zero tooling requirement
  - but support go and docker if installed
- automatic build (on demand)
- no configuration, only the test script
- backwards compatibility (imports, registry)
- support version constraints
- ready for cloud
- JavaScript-like behavior

---

## How It Works

1. test script analysis
   - recursively following local imports
2. creating a list of extensions
   - with optional version constraints
3. using cached k6 binary if possible
4. resolving extension names into go modules
5. resolving version constraints
6. build k6 binary and cache
   - builders: service, native, docker
---

### Builder Service

- HTTP service with GET method
  - the path identifies the k6 binary
  - /os/arch/sorted-versioned-dependency-list
  - each k6 binary is built only once
  - response is cacheable forever
- local HTTP accelerator proxy and edge cache
- with persistent go build (and source) cache
  - go build is really fast
- preload feature for go build cache

---

### Docker Builder

- persistent cache volume
- almost as fast as the native go build
- custom docker image (`szkiba/k6x`)
- remote docker engine support
  - also via ssh
- no docker CLI required to use it

---

### Native Builder

- using locally installed go tooling
- fast due to local go build cache
  - although the service builder is often faster
- the `szkiba/k6x` docker image uses it by default

---

# Proposals

Additional k6 extension ecosystem tools

---

### xk6-lint

*k6 extension linter tool (proposal)*

- check common extension requirements
  - naming, verisoning, README.md, etc
- available as CLI tool, HTTP service and GH action
- grades by compliance level (A+,A,B,C,D,E,F)
- service
  - check can run automatically or on trigger
  - provides colored badges ![xk6 comliance](https://img.shields.io/badge/xk6_compliance-A%2B-green) .. ![xk6 comliance](https://img.shields.io/badge/xk6_compliance-D-yellow) .. ![xk6 comliance](https://img.shields.io/badge/xk6_compliance-F-red)
  - can be fed by GitHub(GitLab) `xk6` topic search
  - can be used as self service extension registry

---

### xk6-kickstart

*k6 extension scaffolding tool (proposal)*

- generate extension skeleton
  - based on extension type (JavaScript, Output)
- updates the extension's source code
  - recommended k6 version in go.mod
  - skeleton change follow-up
  - tries to handle incompatible k6 API changes
- available as CLI and GitHub CLI (gh) extension
  - and as GitHub dependabot like GH action

---

That's All Folks!

