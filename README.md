<!-- template:define:options
{
  "nodescription": true
}
-->
![logo](https://liam.sh/-/gh/svg/lrstanley/go-ytdlp?layout=left&icon=logos%3Ayoutube-icon&icon.height=70&font=1.2&bg=geometric&bgcolor=rgba%2833%2C+33%2C+33%2C+1%29)

<!-- template:begin:header -->
<!-- do not edit anything in this "template" block, its auto-generated -->

<p align="center">
  <a href="https://github.com/lrstanley/go-ytdlp/tags">
    <img title="Latest Semver Tag" src="https://img.shields.io/github/v/tag/lrstanley/go-ytdlp?style=flat-square">
  </a>
  <a href="https://github.com/lrstanley/go-ytdlp/commits/master">
    <img title="Last commit" src="https://img.shields.io/github/last-commit/lrstanley/go-ytdlp?style=flat-square">
  </a>


  <a href="https://github.com/lrstanley/go-ytdlp/actions?query=workflow%3Atest+event%3Apush">
    <img title="GitHub Workflow Status (test @ master)" src="https://img.shields.io/github/actions/workflow/status/lrstanley/go-ytdlp/test.yml?branch=master&label=test&style=flat-square">
  </a>




  <a href="https://codecov.io/gh/lrstanley/go-ytdlp">
    <img title="Code Coverage" src="https://img.shields.io/codecov/c/github/lrstanley/go-ytdlp/master?style=flat-square">
  </a>

  <a href="https://pkg.go.dev/github.com/lrstanley/go-ytdlp">
    <img title="Go Documentation" src="https://pkg.go.dev/badge/github.com/lrstanley/go-ytdlp?style=flat-square">
  </a>
  <a href="https://goreportcard.com/report/github.com/lrstanley/go-ytdlp">
    <img title="Go Report Card" src="https://goreportcard.com/badge/github.com/lrstanley/go-ytdlp?style=flat-square">
  </a>
</p>
<p align="center">
  <a href="https://github.com/lrstanley/go-ytdlp/issues?q=is:open+is:issue+label:bug">
    <img title="Bug reports" src="https://img.shields.io/github/issues/lrstanley/go-ytdlp/bug?label=issues&style=flat-square">
  </a>
  <a href="https://github.com/lrstanley/go-ytdlp/issues?q=is:open+is:issue+label:enhancement">
    <img title="Feature requests" src="https://img.shields.io/github/issues/lrstanley/go-ytdlp/enhancement?label=feature%20requests&style=flat-square">
  </a>
  <a href="https://github.com/lrstanley/go-ytdlp/pulls">
    <img title="Open Pull Requests" src="https://img.shields.io/github/issues-pr/lrstanley/go-ytdlp?label=prs&style=flat-square">
  </a>
  <a href="https://github.com/lrstanley/go-ytdlp/discussions/new?category=q-a">
    <img title="Ask a Question" src="https://img.shields.io/badge/support-ask_a_question!-blue?style=flat-square">
  </a>
  <a href="https://liam.sh/chat"><img src="https://img.shields.io/badge/discord-bytecord-blue.svg?style=flat-square" title="Discord Chat"></a>
</p>
<!-- template:end:header -->

<!-- template:begin:toc -->
<!-- do not edit anything in this "template" block, its auto-generated -->
## :link: Table of Contents

  - [Features](#sparkles-features)
    - [Help Documentation Example](#sparkles-help-documentation-example)
  - [Usage](#gear-usage)
  - [Examples](#clap-examples)
    - [Simple](#simple)
  - [Support &amp; Assistance](#raising_hand_man-support--assistance)
  - [Contributing](#handshake-contributing)
  - [License](#balance_scale-license)
<!-- template:end:toc -->

## :sparkles: Features

**!!! NOTE: go-ytdlp isn't stable yet, and as such, there may be wide-reaching _breaking_ changes,
until 1.0.0 !!!**

- CLI bindings for yt-dlp -- including all flags/commands.
- Optional `Install` and `MustInstall` helpers to auto-download the latest supported version of
  yt-dlp, including proper checksum validation for secure downloads.
  - Worry less about making sure yt-dlp is installed wherever **go-ytdlp** is running from!
- Carried over help documentation for all functions/methods.
- Flags with arguments have type mappings according to what the actual flags expect.
- Completely generated, ensuring it's easy to update to future **yt-dlp** versions.
- Deprecated flags are marked as deprecated in a way that should be caught by most IDEs/linters.
- Stdout/Stderr parsing, with timestamps, and optional JSON post-processing.

### :sparkles: Help Documentation Example

![help documentation example](https://cdn.liam.sh/share/2023/09/Code_m1wz0zsCj9.png)

---

## :gear: Usage

<!-- template:begin:goget -->
<!-- do not edit anything in this "template" block, its auto-generated -->
```console
go get -u github.com/lrstanley/go-ytdlp@latest
```
<!-- template:end:goget -->

## :clap: Examples

### Simple

See also [_examples/simple/main.go](./_examples/simple/main.go), which includes
writing results (stdout/stderr/etc) as JSON.

```go
package main

import (
	"context"

	"github.com/lrstanley/go-ytdlp"
)

func main() {
	// If yt-dlp isn't installed yet, download and cache it for further use.
	ytdlp.MustInstall(context.TODO(), nil)

	dl := ytdlp.New().
		FormatSort("res,ext:mp4:m4a").
		RecodeVideo("mp4").
		Output("%(extractor)s - %(title)s.%(ext)s")

	_, err := dl.Run(context.TODO(), "https://www.youtube.com/watch?v=dQw4w9WgXcQ")
	if err != nil {
		panic(err)
	}
}
```

---

<!-- template:begin:support -->
<!-- do not edit anything in this "template" block, its auto-generated -->
## :raising_hand_man: Support & Assistance

* :heart: Please review the [Code of Conduct](.github/CODE_OF_CONDUCT.md) for
     guidelines on ensuring everyone has the best experience interacting with
     the community.
* :raising_hand_man: Take a look at the [support](.github/SUPPORT.md) document on
     guidelines for tips on how to ask the right questions.
* :lady_beetle: For all features/bugs/issues/questions/etc, [head over here](https://github.com/lrstanley/go-ytdlp/issues/new/choose).
<!-- template:end:support -->

<!-- template:begin:contributing -->
<!-- do not edit anything in this "template" block, its auto-generated -->
## :handshake: Contributing

* :heart: Please review the [Code of Conduct](.github/CODE_OF_CONDUCT.md) for guidelines
     on ensuring everyone has the best experience interacting with the
    community.
* :clipboard: Please review the [contributing](.github/CONTRIBUTING.md) doc for submitting
     issues/a guide on submitting pull requests and helping out.
* :old_key: For anything security related, please review this repositories [security policy](https://github.com/lrstanley/go-ytdlp/security/policy).
<!-- template:end:contributing -->

<!-- template:begin:license -->
<!-- do not edit anything in this "template" block, its auto-generated -->
## :balance_scale: License

```
MIT License

Copyright (c) 2023 Liam Stanley <liam@liam.sh>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

_Also located [here](LICENSE)_
<!-- template:end:license -->
