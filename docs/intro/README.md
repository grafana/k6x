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
We see this redirect in the output of curl as the first response.
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

<!--
Well, what is k6x then?

First of all, a k6 launcher, which ensures that the k6 binary will contain the extensions used by the test script.

It is available as a native binary or as a docker image.

No additional configuration is required to use it, only the test script.

There is no need to install additional tools to use it.
However, it supports the use of go and docker engine (even remote) if it is installed.

It can be used as k6 builder tool, native or docker based.

It can also be used as a k6 builder or download service.
-->

---

# Under The Hood

<!--
A few technological details about the k6x follow.
-->

---

## Design Considerations

- zero tooling requirement
  - but support go and docker if installed
- automatic build (on demand)
- no additional configuration, only the test script
- backwards compatibility (imports, registry)
- support version constraints
- ready for cloud
- JavaScript-like behavior

<!--
What were the design considerations of k6x?

One of the most important design goals was to make it easy to use, without the need to install any tools.

k6 should be built or downloaded automatically based on the test script.

It must be backwards compatible, it cannot require changes to the test script that prevent it from running with standard k6.
The k6 extension registry should be considered as the primary source of truth.

Since extensions change over time, it should be possible to use version constraints, but optionally.

It should be suitable for use in a cloud environment.

Since the tests are written in JavaScript, solutions common in the JavaScript ecosystem should be preferred (use pragma, version constraints format)
-->

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

<!--
A few bullet points about how k6x works. A more detailed description can be found in the readme file.

The first step is to analyze the test script, recursively following the local imports.

A list of required extensions is then created, taking into account optional version constraints.

If the previously used k6 binary found in the cache contains the necessary extensions, it will be used.

Otherwise, previously used extensions are added to the list. This is necessary to make the local cache more efficient.

The names of the extensions are resolved to the go module names using the extension registry.

The go module versions are selected taking into account the version constraints.

The k6 binary is built using the appropriate builder. The default builder precedence is based on speed.

It will use the first available builder.
-->

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

<!--
The builder service is based on the HTTP GET method. Therefore, it can be considered both a builder service and a download service. The k6 binary is built only on the first call, after which it is downloaded from the cache.

The GET method path clearly identifies the k6 binary. It includes the operating system, processor architecture, and all dependencies with version numbers. Of course, the dependencies also include k6, with a version number. k6 is the first item in the dependency list, followed by other dependencies in alphabetical order.

Since the path clearly identifies the k6 binary, the answer can be cached forever.

In front of the service there is a local HTTP accelerator proxy and an edge cache.

The go build cache is persistent and stores the partial results of compiling each extension. Therefore, the occasional build is also fast.

When new versions are released, a cache preload command can be run, which compiles the latest version of all extensions and k6, thereby preloading the go build cache.
-->

---

### Builder Service Availability

- no public builder service (yet)
  - only `native` and `docker` builder can be used
  - or a self-hosted builder service
- there is a builder service for closed beta testing
- set `K6X_BUILDER_SERVICE` to builder service URL

<!--
There is currently no public build service. Only native and docker builder can be used. Or you can run your own build service, simply start the k6x binary in service mode.

For development, I created a build service instance under my own Oracle Cloud account, in the free tier. I connected a CloudFlare edge cache in front of it, also in the free tier. It works at an acceptable speed for personal use, but of course it is not suitable for beta testing.

I created a builder service instance on an AWS virtual machine for beta testing. I will share its address on the discussion chat channel and you will be able to try it out. In fact, I'd like to ask you to try it, I'm looking forward to any feedback.

To use the builder service, you need to set the K6X_BUILDER_SERVICE environment variable to the URL of the builder service.

Of course, once there is a publicly available builder service, its address will be set by default.
-->

---

### Docker Builder

- persistent cache volume
- almost as fast as the native go build
- custom docker image (`szkiba/k6x`)
- remote docker engine support
  - also via ssh
- no docker CLI required to use it

<!--

To use docker builder, you only need a docker engine, even a remote one.

k6x communicates directly with the docker engine, so no docker CLI is required. Remote docker engine can be used via TCP or SSH protocol.

For the build, k6x will automatically create a persistent cache volume and store the go build cache there. Therefore, the docker builder is almost as fast as the native go builder.
-->

---

### Native Builder

- using locally installed go tooling
- fast due to local go build cache
  - although the service builder is often faster
- the `szkiba/k6x` docker image uses it by default

<!--
The native builder can be used if you have a locally installed go compiler and git. There is no need to install xk6, k6x includes xk6 as a built-in library.

The native builder is fast, thanks to the local go build cache. Although the service builder is often faster, especially if the k6 binary is already in the HTTP cache.

The k6x docker image uses this builder by default because the docker image includes a go compiler.
-->

---

### xk6-lint

*k6 extension linter tool (proposal)*

- check common extension requirements
  - naming, versioning, README.md, etc
- available as CLI tool, HTTP service and GH action
- grades by compliance level (A+,A,B,C,D,E,F)
- service
  - check can run automatically or on trigger
  - provides colored badges ![xk6 comliance](https://img.shields.io/badge/xk6_compliance-A%2B-green) .. ![xk6 comliance](https://img.shields.io/badge/xk6_compliance-D-yellow) .. ![xk6 comliance](https://img.shields.io/badge/xk6_compliance-F-red)
  - can be fed by GitHub(GitLab) `xk6` topic search
  - can be used as self service extension registry

<!--
Finally, a tool development proposal.

Using extensions requires trust from users. It would be nice to see how each extension meets expectations. The usual solution for this is the use of so-called linter tools.

An extension linter tool would also be suitable for registry maintenance. A linter service could even replace the extension registry.

Main features:

It would check how well the extension meets the requirements of the specification. For example: is the repository named correctly, is the extension versioned correctly, is there a readme file, is the extension using the correct k6 version, and so on.

It would be classified in an easily understandable grade based on compliance with the specification.
A+, A, B, C and so on.

The linter service would provide easily recognizable colored badges based on the grades. Badges are good, everyone understands them, everyone loves them. With the badges, users can easily verify the quality of the given extension.

The check can be scheduled or triggered.

The linter service can also be considered a self-service extension registry. Developers can initiate the verification using a GitHub action, which also registers their extension.
-->

---

That's All Folks!

<!--
That's all I wanted to share with you about the k6x in brief. I hope this is just the beginning of the journey.
-->
