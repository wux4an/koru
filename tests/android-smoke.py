#!/usr/bin/env python3
"""Real Android WebView tests via its debugging port, never a build dependency.
Requires an already installed debug APK and `adb forward tcp:9223 ...`.
"""

import asyncio, json, time, urllib.request, websockets, argparse, pathlib


async def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--port", type=int, default=9223)
    parser.add_argument("--output", default="docs/results/full/android-smoke.json")
    args = parser.parse_args()
    base = f"http://127.0.0.1:{args.port}"
    pages = json.load(urllib.request.urlopen(base + "/json"))
    page = next(p for p in pages if p["type"] == "page")
    async with websockets.connect(page["webSocketDebuggerUrl"], max_size=2**22) as ws:
        counter = 0

        async def command(method, params):
            nonlocal counter
            counter += 1
            id = counter
            await ws.send(json.dumps({"id": id, "method": method, "params": params}))
            while True:
                event = json.loads(await ws.recv())
                if event.get("id") == id:
                    if "error" in event:
                        raise RuntimeError(event["error"])
                    return event["result"]

        async def evaluate(expression):
            result = await command(
                "Runtime.evaluate",
                {"expression": expression, "awaitPromise": True, "returnByValue": True},
            )
            if "exceptionDetails" in result:
                raise AssertionError(result["exceptionDetails"])
            return result["result"].get("value")

        await command("Page.enable", {})
        for _ in range(60):
            if await evaluate('!!document.body && typeof Koru === "object"'):
                break
            await asyncio.sleep(1)
        else:
            raise AssertionError("web document not ready")
        expression = r"""(async()=>{
   const results={url:location.href,title:document.title};
   function assert(condition,message){if(!condition)throw Error(message);}
   await Promise.race([Koru.ready,new Promise((_,reject)=>setTimeout(()=>reject(Error('channel not ready')),10000))]);
   const start=performance.now();results.device=await Koru.invoke('device.info');results.bridgeMillis=performance.now()-start;
   assert(results.device.platform==='android','device platform');results.app=await Koru.invoke('app.info');assert(results.app.appId==='com.example.koru','app identity');
   await Koru.invoke('storage.set',{key:'smoke',value:'persisted'});results.storage=await Koru.invoke('storage.get',{key:'smoke'});assert(results.storage.value==='persisted','storage round trip');
   await Koru.invoke('storage.remove',{key:'smoke'});const missing=await Koru.invoke('storage.get',{key:'smoke'});assert(missing.value===null,'removed storage null');
   await Koru.invoke('storage.set',{key:'restart',value:'survives-restart'});
   results.native=await Koru.invoke('native.info');assert(results.native.runtime==='JNI'&&results.native.abi===1&&results.native.thread==='main','JNI runtime');
   results.vibration=await Koru.invoke('vibration.vibrate',{duration:20});
   results.notificationPermission=await Koru.invoke('notifications.permission');
   results.notification=await Koru.invoke('notifications.show',{id:17,title:'Koru local',body:'Offline notification'});
   results.errorCodes=[];
   for(const [method,params] of [['missing.method',{}],['storage.set',{key:'safe',value:123}],['vibration.vibrate',{duration:-1}]]){
    try{await Koru.invoke(method,params);throw Error('invalid request succeeded');}catch(e){assert(e.code==='METHOD_NOT_FOUND'||e.code==='INVALID_REQUEST','structured errors');results.errorCodes.push(e.code);}
   }
   const button=document.querySelector('#count');button.click();assert(button.textContent==='Contador: 1','JavaScript counter');
   results.css=getComputedStyle(document.body).padding;assert(results.css!=='0px','CSS loaded');
   const image=document.querySelector('img');assert(image.complete&&image.naturalWidth>0,'local image loaded');results.imageWidth=image.naturalWidth;
   results.module=await import('/__koru/core.mjs').then(m=>m.device.getInfo());assert(results.module.platform==='android','ES module API');
   const forbidden=await fetch('/not-a-resource.txt');assert(forbidden.status===404,'unknown resource isolation');results.resourceStatus=forbidden.status;
   results.origin=location.origin;assert(results.origin==='https://app.koru.invalid','controlled HTTPS origin');
   return results;
  })()"""
        results = await evaluate(expression)
        pathlib.Path(args.output).write_text(json.dumps(results, indent=2) + "\n")
        print(json.dumps(results, indent=2))


asyncio.run(main())
