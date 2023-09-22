"use strict";

const Hashira = {
    Fetch: (wasmURL) => {
        return new Promise((resolve, reject) => {
            if (!WebAssembly.instantiateStreaming) {
                WebAssembly.instantiateStreaming = async (resp, importObject) => {
                    const source = await (await resp).arrayBuffer();
                    return await WebAssembly.instantiate(source, importObject);
                };
            }

            let go = new Go();
            WebAssembly.instantiateStreaming(fetch(wasmURL), go.importObject).then(
                (result) => {
                    go.run(result.instance)
                    // TODO: in the future this should be instance based
                    // once multiple instances are supported by Hashira
                    resolve(new HashiraClient());
                }
            );
        });
    }
};

class HashiraClient {
    constructor() {
    }

    bindCanvasByID(canvasID) {
        window.HashiraInitRenderLoop(canvasID);
    }

    sendEvent(event, data) {
        window.HashiraSendEvent(event, data);
    }

    loadTileset(url, tileSize) {
        fetch(url).then((response) => {
            return response.arrayBuffer();
        }).then((buffer) => {
            this.sendEvent("resources.LoadTileset", { tileSize: tileSize, data: new Uint8Array(buffer) });

        });
    }

    setBackgroundColor(hex) {
        this.sendEvent("world.SetBackground", { color: hex });
    }

    addMap(name, width, height) {
        this.sendEvent("world.AddMap", { name: name, width: width, height: height });
    }

    addLayer(mapName, layerName, z) {
        this.sendEvent("world.AddLayer", { map: mapName, name: layerName, z: z });
    }

    addLayerData(mapName, layerName, data) {
        this.sendEvent("world.AddLayerData", { map: mapName, name: layerName, data: data });
    }

    setCameraZoom(zoom) {
        this.sendEvent("camera.Zoom", { zoom: zoom });
    }

    setCameraTranslation(x, y) {
        this.sendEvent("camera.Translate", { x: x, y: y });
    }

    setCameraToMapCenter(mapName) {
        this.sendEvent("camera.TranslateToMapCenter", { map: mapName });
    }
}