#!/usr/bin/env python3
"""Small reusable QA helper. Executes only in the installed developer test app."""

import asyncio, json, urllib.request, websockets, sys


async def main():
    pages = json.load(urllib.request.urlopen("http://127.0.0.1:9223/json"))
    p = next(x for x in pages if json.loads(x["description"]).get("visible"))
    async with websockets.connect(p["webSocketDebuggerUrl"]) as ws:
        await ws.send(
            json.dumps(
                {
                    "id": 1,
                    "method": "Runtime.evaluate",
                    "params": {
                        "expression": sys.argv[1],
                        "awaitPromise": True,
                        "returnByValue": True,
                    },
                }
            )
        )
        while True:
            d = json.loads(await ws.recv())
            if d.get("id") == 1:
                print(json.dumps(d, indent=2))
                break


asyncio.run(main())
