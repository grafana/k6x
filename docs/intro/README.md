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

<!--
Thank you for watching the presentation about k6x.
The primary goal of k6x is to make it easier to use k6 with extensions.
k6x is currently a pet project, but I hope it will become an official project in the near future.
Let's see what this presentation will be about.
-->

---

# Agenda

- Use Cases
- Under The Hood
- xk6-lint

<!--
Before we get to what the k6x is, I'll show you what it can be used for through a few use cases.
Then I'll show you what's under the hood.
Finally, I will introduce a new tool proposal that can be used to measure and improve the quality of extensions.
-->

---

# Use Cases

What is k6x?

<!--
The following use cases will be familiar if you have tried using an extension with k6.
I will show you what the k6x is good for.
-->

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

<!--
The first use case is the most common. I just want to run the test.
I don't want to know which extension is located where, how to build k6 with it.
I just want to run the test. I just want to use the extensions.
It's as simple as that, what I want.

I will use this example test. Let's imagine that we have an encrypted YAML file in Ansible Valult format.
We expect the encryption key in an environment variable, which is set in a file named .env and not stored in git.

The YAML file contains api keys for a service.
We want to call the service with randomly generated data using the api key.

It wasn't easy to use 4 extensions in one screen of example code, but I did it.

So we would like to run this test.
-->

---

> I just want to run my test!

```bash
k6x run script.js
```

![script.js](script.svg)

<!--
Well, all we need to do is run the test using the k6x command instead of the k6 command.

It's that simple, k6x does the rest.

The first time it takes a minimal amount of time to provide the correct k6 binary, but the second run happens immediately because of the cache.
-->

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

<!--
Extensions change from time to time, new functions are added to the API, or function parameters change.
This means that our test may only work with a certain version of the extension. In such cases, it is useful to be able to specify version constraints.

Version constraints can be specified with the "use k6" pragma, similar to the "use strict" pragma.
Using version constraints is optional.
-->
---

> I want to run my test, but with...

```bash
k6x run script-with.js
```

![script-with.js](script-with.svg)

<!--
k6x takes version constraints into account when running the test and will use the appropriate version of each extension.
-->
---

> I want to run my test, but with output extension

```bash
k6x run -o top -d 5s script.js
```

![top.js](top.svg)

<!--
In addition to the extensions used in the test, k6x automatically detects the use of output extensions.
In this example, we use the xk6-top output extension to display metrics in the terminal.
-->

---

> I want to use k6 extensions in the cloud

```bash
docker run --rm -it -e K6X_BUILDER_SERVICE=$K6X_BUILDER_SERVICE -v $PWD:/home/k6x szkiba/k6x run script.js
```

![cloud](cloud.svg)

<!--
It is a natural requirement that we want to use the extensions in a cloud environment.
In this example, k6x is running in a container as if it were running in a cloud environment.
-->

---

> I want to control which extensions are allowed

```bash
k6x run --filter "[?contains(tiers,'Official')]" script.js
```

![filter](filter.svg)

<!--
There are times when the use of an arbitrary extension is not allowed.
It may be necessary to control the allowed extensions.
The k6x makes this possible by using a filter.
You can filter on any property of the extension.

In this example, only official extensions are allowed.
That's why we get an error message.
-->

---

> I need a k6 binary with extensions

```bash
curl -OJ $K6X_BUILDER_SERVICE/linux/amd64/k6@v0.47.0,dashboard@v0.6.0,k6/x/faker@v0.2.2,top@v0.1.1
chmod +x k6
./k6 version
```

![get](get.svg)

<!--
Sometimes I don't want to use k6x, I just need a k6 binary with extensions.
In this case, the appropriate k6 binary can be downloaded from the builder service using any HTTP client.
The exact version of each extension must be specified in the path.
-->

---

> I need a k6 binary with extensions, but I'm lazy

```bash
curl -OJL $K6X_BUILDER_SERVICE/linux/amd64/dashboard,top,k6/x/faker
chmod +x k6
./k6 version
```

![get-lazy](get-lazy.svg)

<!--
The previous URL was quite long and I'm quite lazy.
The builder service also supports a looser definition, in which case it resolves the missing versions to the latest available.

After resolving the versions, a redirect response will be sent to the URL containing the exact versions.
-->

---

> I want to build k6 binary with extensions

```bash
k6x build --with dashboard --with k6/x/faker --with top --with k6/x/yaml
./k6 version
```

![build](build.svg)

<!--
k6x can also be used as a build tool.
I can specify which extensions I need on the command line or I can specify a test and it will detect which extensions I need.
And it makes the corresponding k6 binary.
-->

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

### Builder Service Availability

- no public builder service (yet)
  - only `native` and `docker` builder can be used
  - or a self-hosted builder service
- there is a builder service for closed beta testing
- set `K6X_BUILDER_SERVICE` to builder service URL

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

That's All Folks!

