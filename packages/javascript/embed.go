// Package javascript embeds the browser bridge; runtime builds need no Node/npm.
package javascript

import _ "embed"

//go:embed core.js
var Core []byte

//go:embed core.mjs
var Module []byte
