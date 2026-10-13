#!/usr/bin/env python3
import asyncio, json, urllib.request, websockets, pathlib


async def main():
    pages = json.load(urllib.request.urlopen("http://127.0.0.1:9223/json"))
    p = next(x for x in pages if json.loads(x["description"]).get("visible"))
    async with websockets.connect(p["webSocketDebuggerUrl"]) as ws:
        expr = """(async()=>{const result={secure:isSecureContext};const registration=await navigator.serviceWorker.register('/sw.js');await navigator.serviceWorker.ready;if(!navigator.serviceWorker.controller)await new Promise(resolve=>navigator.serviceWorker.addEventListener('controllerchange',resolve,{once:true}));result.scope=registration.scope;result.probe=await fetch('/sw-probe').then(r=>r.text());if(result.probe!=='koru-worker-offline')throw Error('worker interception failed');return result})()"""
        await ws.send(
            json.dumps(
                {
                    "id": 1,
                    "method": "Runtime.evaluate",
                    "params": {
                        "expression": expr,
                        "awaitPromise": True,
                        "returnByValue": True,
                    },
                }
            )
        )
        while True:
            d = json.loads(await asyncio.wait_for(ws.recv(), 60))
            if d.get("id") == 1:
                if "exceptionDetails" in d.get("result", {}):
                    raise AssertionError(d)
                value = d["result"]["result"]["value"]
                print(json.dumps(value, indent=2))
                pathlib.Path("docs/results/full/android-worker.json").write_text(
                    json.dumps(value, indent=2) + "\n"
                )
                break


asyncio.run(main())
