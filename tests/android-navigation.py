#!/usr/bin/env python3
import asyncio, json, urllib.request, websockets, subprocess, pathlib

import os

adb = os.environ.get("KORU_TEST_ADB", "adb")


async def main():
    p = json.load(urllib.request.urlopen("http://127.0.0.1:9223/json"))[0]
    results = {}
    async with websockets.connect(p["webSocketDebuggerUrl"]) as ws:
        n = 0

        async def ev(expr):
            nonlocal n
            n += 1
            await ws.send(
                json.dumps(
                    {
                        "id": n,
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
                d = json.loads(await ws.recv())
                if d.get("id") == n:
                    if "exceptionDetails" in d["result"]:
                        raise AssertionError(d)
                    return d["result"]["result"].get("value")

        for _ in range(30):
            if await ev("!!document.querySelector('a')"):
                break
            await asyncio.sleep(1)
        r = await ev(
            "(()=>{const r=document.querySelector('a').getBoundingClientRect();return{x:(r.left+r.width/2)*devicePixelRatio,y:(r.top+r.height/2)*devicePixelRatio+48}})()"
        )
        subprocess.run(
            [adb, "shell", "input", "tap", str(round(r["x"])), str(round(r["y"]))],
            check=True,
        )
        await asyncio.sleep(5)
        results["navigation"] = await ev("location.href")
        assert results["navigation"].endswith("/second.html"), results
        subprocess.run([adb, "shell", "input", "keyevent", "4"], check=True)
        await asyncio.sleep(5)
        # WebView may replace its DevTools target during a history restoration.
        try:
            results["back"] = await ev("location.href")
        except websockets.exceptions.ConnectionClosed:
            pages = json.load(urllib.request.urlopen("http://127.0.0.1:9223/json"))
            active = next(
                x for x in pages if json.loads(x["description"]).get("visible")
            )
            results["back"] = active["url"]
            results["devtoolsTargetReplaced"] = True
        assert results["back"].endswith("/index.html"), results
        active = json.load(urllib.request.urlopen("http://127.0.0.1:9223/json"))[0]
        async with websockets.connect(active["webSocketDebuggerUrl"]) as restored:
            await restored.send(
                json.dumps(
                    {
                        "id": 99,
                        "method": "Runtime.evaluate",
                        "params": {
                            "expression": "Koru.invoke('storage.get',{key:'restart'})",
                            "awaitPromise": True,
                            "returnByValue": True,
                        },
                    }
                )
            )
            while True:
                reply = json.loads(await restored.recv())
                if reply.get("id") == 99:
                    results["storage"] = reply["result"]["result"]["value"]
                    break
        assert results["storage"]["value"] == "survives-restart", results
    print(json.dumps(results, indent=2))
    pathlib.Path("docs/results/full/android-back.json").write_text(
        json.dumps(results, indent=2) + "\n"
    )


asyncio.run(main())
