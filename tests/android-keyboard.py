#!/usr/bin/env python3
"""Exercise real Android text input and Back; reconnect after WebView resize."""

import asyncio
import json
import os
import pathlib
import subprocess
import urllib.request

import websockets

adb = os.environ.get("KORU_TEST_ADB", "adb")


async def evaluate(expression):
    # Showing/hiding the IME may replace the DevTools target without reloading DOM.
    for _ in range(20):
        pages = json.load(urllib.request.urlopen("http://127.0.0.1:9223/json"))
        page = next(p for p in pages if json.loads(p["description"]).get("visible"))
        try:
            async with websockets.connect(page["webSocketDebuggerUrl"]) as ws:
                await ws.send(
                    json.dumps(
                        {
                            "id": 1,
                            "method": "Runtime.evaluate",
                            "params": {"expression": expression, "returnByValue": True},
                        }
                    )
                )
                while True:
                    result = json.loads(await ws.recv())
                    if result.get("id") == 1:
                        assert "exceptionDetails" not in result["result"], result
                        return result["result"]["result"].get("value")
        except websockets.exceptions.ConnectionClosed:
            await asyncio.sleep(0.5)
    raise AssertionError("WebView DevTools target did not stabilize")


async def main():
    subprocess.run(
        [adb, "shell", "settings", "put", "secure", "show_ime_with_hard_keyboard", "1"],
        check=True,
    )
    bounds = await evaluate(
        "(()=>{const old=document.getElementById('keyboard');if(old)old.remove();const i=document.createElement('input');i.id='keyboard';i.placeholder='Keyboard test';document.body.prepend(i);const r=i.getBoundingClientRect();return{x:(r.x+r.width/2)*devicePixelRatio,y:(r.y+r.height/2)*devicePixelRatio+48}})()"
    )
    subprocess.run(
        [
            adb,
            "shell",
            "input",
            "tap",
            str(round(bounds["x"])),
            str(round(bounds["y"])),
        ],
        check=True,
    )
    await asyncio.sleep(3)
    subprocess.run([adb, "shell", "input", "text", "KoruKeyboard"], check=True)
    await asyncio.sleep(2)
    result = await evaluate(
        "({value:document.getElementById('keyboard').value,focused:document.activeElement.id,height:innerHeight})"
    )
    assert result["value"] == "KoruKeyboard", result
    ime = subprocess.check_output([adb, "shell", "dumpsys", "input_method"], text=True)
    result["keyboardShown"] = "mInputShown=true" in ime
    assert result["keyboardShown"], result
    subprocess.run([adb, "shell", "input", "keyevent", "4"], check=True)
    await asyncio.sleep(2)
    result["afterBackURL"] = await evaluate("location.href")
    assert result["afterBackURL"].startswith("https://app.koru.invalid/"), result
    ime = subprocess.check_output([adb, "shell", "dumpsys", "input_method"], text=True)
    result["keyboardHiddenAfterBack"] = "mInputShown=false" in ime
    assert result["keyboardHiddenAfterBack"], result
    pathlib.Path("docs/results/full/android-keyboard.json").write_text(
        json.dumps(result, indent=2) + "\n"
    )
    print(json.dumps(result, indent=2))


asyncio.run(main())
