// Same-origin worker served from the app's assets.
self.onmessage = (event) => self.postMessage({ echo: event.data, secure: self.isSecureContext });
