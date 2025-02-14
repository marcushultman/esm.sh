# cjs-esm-exports

A **WASM** module to parse commonjs exports for **ESM**, powered by [swc](https://github.com/swc-project/swc) in **rust**.

## Installation

```bash
npm install cjs-esm-exports
```

for `yarn` users:

```bash
yarn add cjs-esm-exports
```

## Usage

Types:
```ts
export function parse(
  specifier: string,
  code: string,
  node_env?: 'development' | 'production',
  call_mode?: boolean,
): {
  exports: string[],
  reexports: string[],
};
```

Example: 
```js
const { parse } = require('cjs-esm-exports');

// named exports
// exports: ['a', 'b', 'c', '__esModule', 'foo']
const { exports } = parse('index.cjs', `
  /* exports.ignore = "not detected"; */
  exports.a = "a";
  module.exports.b = "b";
  Object.defineProperty(exports, "c", { value: "c" });
  Object.defineProperty(module.exports, "__esModule", { value: true })

  const key = "foo"
  Object.defineProperty(exports, key, { value: "e" });
`);

// reexports
// reexports: ['./lib']
const { reexports } = parse('index.cjs', `
  module.exports = require("./lib");
`);

// object exports(spread supported)
// exports: ['foo', 'baz']
// reexports: ['./lib']
const { reexports } = parse('index.cjs', `
  const foo = 'bar'
  const obj = { baz: 123 }
  module.exports = { foo, ...obj, ...require("./lib") };
`);

// condition
// exports: ['foo', 'cjs']
const { exports } = parse('index.cjs', `
  module.exports.a = "a";
  if (true) {
    exports.foo = "bar";
  }
  const mtype = "cjs";
  if (mtype === "cjs") {
    exports.cjs = true;
  } else {
    exports.esm = true;
  }
  if (false) {
    exports.ignore = "ignore";
  }
`);

// block&IIFE
// exports: ['foo', 'baz', '__esModule']
const { exports } = parse('index.cjs', `
  (function () {
    exports.foo = "bar"
    if (true) {
      return
    }
    exports.ignore = '-'
  })();
  {
    exports.baz = 123
  }
  exports.__esModule = true
`);

// condition with `process.env.NODE_ENV`
// reexports: ['./index.development']
const { reexports } = parse('index.cjs', `
  if (process.env.NODE_ENV === "development") {
    module.exports = require("./index.development")
  } else {
    module.exports = require("./index.production")
  }
`, 'development');

// IIFE exports
// exports: ['foo']
const { exports } = parse('index.cjs', `
  function Fn() {
    return { foo: "bar" }
  }
  module.exports = Fn()
`);

// function reexports
// reexports: ['./lib()']
const { reexports } = parse('index.cjs', `
  module.exports = require("./lib")()
`);

// apply function exports (call mode)
// exports: ['foo']
const { reexports } = parse('lib.cjs', `
  module.exports = function() {
    return { foo: 'bar' }
  }
`, 'production', true);
```

## Development Setup

You will need [rust](https://www.rust-lang.org/tools/install) 1.30+ and [wasm-pack](https://rustwasm.github.io/wasm-pack/installer/).

## Build

```bash
wasm-pack build --target nodejs
```

## Run tests

```bash
cargo test --all
```
