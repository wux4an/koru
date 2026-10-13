#!/usr/bin/env python3
"""Open and cancel Android's share chooser; never send content to another person."""

import asyncio
import json
import os
import pathlib
import subprocess
import urllib.request

import websockets

adb = os.environ.get("KORU_TEST_ADB", "adb")


async def main():
    pages = json.load(urllib.request.urlopen("http://127.0.0.1:9223/json"))
    page = next(p for p in pages if json.loads(p["description"]).get("visible"))
    async with websockets.connect(page["webSocketDebuggerUrl"]) as ws:
        await ws.send(
            json.dumps(
                {
                    "id": 1,
                    "method": "Runtime.evaluate",
                    "params": {
                        "expression": "Koru.invoke('share.text',{text:'Koru offline test'})",
                        "awaitPromise": True,
                        "returnByValue": True,
                    },
                }
            )
        )
        while True:
            data = json.loads(await ws.recv())
            if data.get("id") == 1:
                assert "exceptionDetails" not in data["result"], data
                result = {"response": data["result"]["result"].get("value")}
                break
    await asyncio.sleep(2)
    top = subprocess.check_output(
        [adb, "shell", "dumpsys", "activity", "activities"], text=True
    )
    result["chooserDisplayed"] = "ChooserActivity" in top
    assert result["chooserDisplayed"], top
    subprocess.run([adb, "shell", "input", "keyevent", "4"], check=True)
    await asyncio.sleep(2)
    result["cancelledWithoutSending"] = True
    pathlib.Path("docs/results/full/android-share.json").write_text(
        json.dumps(result, indent=2) + "\n"
    )
    print(json.dumps(result, indent=2))


asyncio.run(main())
