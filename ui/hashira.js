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
            this.sendEvent("ScreenResized", { width: width, height: height });
        }, false);
    }

    bindCanvasByID = (canvasID) => {
        window.HashiraInitRenderLoop(canvasID);
    }

    sendEvent = (event, data) => {
        window.HashiraSendEvent(event, JSON.stringify(data));
    }

    loadTileset = (url) => {
        return fetch(url).then((response) => {
            return response.arrayBuffer();
        }).then((buffer) => {
            this.sendEvent("TilesetLoaded", { bytes: Array.from(new Uint8Array(buffer)) });
        });
    }

    setBackgroundColor = (hex) => {
        this.sendEvent("BackgroundColorSet", { color: hex });
    }

    addMap = (name, width, height, tileWidth, tileHeight) => {
        this.sendEvent("MapAdded", { name: name, width: width, height: height, tile_width: tileWidth, tile_height: tileHeight });
    }

    addLayer = (mapName, layerName, z) => {
        this.sendEvent("LayerAdded", { map: mapName, name: layerName, z: z });
    }

    addLayerData = (mapName, layerName, data) => {
        this.sendEvent("LayerDataAdded", { map: mapName, layer: layerName, data: data });
    }

    setTile = (mapName, layerName, x, y, tileID) => {
        this.sendEvent("TileAssigned", { map: mapName, layer: layerName, x: x, y: y, tile: tileID });
    }

    setCameraZoom = (zoom) => {
        this.sendEvent("CameraZoomed", { zoom: zoom });
    }

    setCameraZoomBy = (by) => {
        this.sendEvent("CammeraZoomedBy", { delta: by });
    }

    setCameraTranslation = (x, y) => {
        this.sendEvent("CameraTranslated", { x: x, y: y });
    }

    setCameraTranslationBy = (x, y) => {
        this.sendEvent("CameraTranslatedBy", { x: x, y: y });
    }

    setCameraToMapCenter = (mapName) => {
        this.sendEvent("CameraTranslatedToMapCenter", { map: mapName });
    }
}
