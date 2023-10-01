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
                    resolve(new HashiraClient(null));
                }
            );
        });
    }
};

class HashiraClient {
    constructor(instance) {
        this.instance = instance;
    }

    bindEvents = (canvas) => {
        this.canvas = canvas;
        window.addEventListener('resize', (e) => {
            const width = this.canvas.clientWidth;
            const height = this.canvas.clientHeight;
            this.sendEvent("screen.Resize", { width: width, height: height });
        }, false);
    }

    bindCanvasByID = (canvasID) => {
        window.HashiraInitRenderLoop(canvasID);
    }

    sendEvent = (event, data) => {
        window.HashiraSendEvent(event, data);
    }

    loadTileset = (url) => {
        return fetch(url).then((response) => {
            return response.arrayBuffer();
        }).then((buffer) => {
            this.sendEvent("resources.LoadTileset", { data: new Uint8Array(buffer) });
        });
    }

    setBackgroundColor = (hex) => {
        this.sendEvent("world.SetBackground", { color: hex });
    }

    addMap = (name, width, height, tileWidth, tileHeight) => {
        this.sendEvent("world.AddMap", { name: name, width: width, height: height, tileWidth: tileWidth, tileHeight: tileHeight });
    }

    addLayer = (mapName, layerName, z) => {
        this.sendEvent("world.AddLayer", { map: mapName, name: layerName, z: z });
    }

    addLayerData = (mapName, layerName, data) => {
        this.sendEvent("world.AddLayerData", { map: mapName, layer: layerName, data: data });
    }

    setTile = (mapName, layerName, x, y, tileID) => {
        this.sendEvent("world.SetTile", { map: mapName, layer: layerName, x: x, y: y, tile: tileID });
    }

    setCameraZoom = (zoom) => {
        this.sendEvent("camera.Zoom", { zoom: zoom });
    }

    setCameraZoomBy = (by) => {
        this.sendEvent("camera.ZoomBy", { delta: by });
    }

    setCameraTranslation = (x, y) => {
        this.sendEvent("camera.Translate", { x: x, y: y });
    }

    setCameraTranslationBy = (x, y) => {
        this.sendEvent("camera.TranslateBy", { x: x, y: y });
    }

    setCameraToMapCenter = (mapName) => {
        this.sendEvent("camera.TranslateToMapCenter", { map: mapName });
    }
}
