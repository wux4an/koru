// Package javascript embeds the browser bridge; runtime builds need no Node/npm.
package javascript

import _ "embed"

//go:embed core.js
var Core []byte

//go:embed core.mjs
var Module []byte

// Types declares the module for TypeScript and editors.
//
//go:embed core.d.ts
var Types []byte
