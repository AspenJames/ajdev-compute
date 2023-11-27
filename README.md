# aspenjames.dev @ edge

Personal, gohtml-based website, deployed to Fastly's [Compute] platform.

Licensed under the [MIT license][LICENSE]

## Depedencies

* [golang]
* [Fastly CLI][fastlycli]
* [Node.js] and [npm]

## Building

The script at `scripts/build.sh` will build the Go WASM particle animation
binary & site CSS, then build these into the Fastly Compute WASM binary. The
`fastly` CLI is configured to execute this build script when running e.g.
`fastly compute serve`. Built artifacts can be cleaned up with
`scripts/clean.sh`.

## Running locally
`fastly compute serve` & have a blast!

## Deploying
TBD

[Compute]: https://docs.fastly.com/products/compute
[fastlycli]: https://developer.fastly.com/learning/tools/cli/
[golang]: https://go.dev/
[LICENSE]: ./LICENSE
[Node.js]: https://nodejs.org/
[npm]: https://www.npmjs.com/
