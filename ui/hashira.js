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
    constructor(canvas) {
        // this.canvas = canvas;
    }

    // bindEvents = () => {
    //     window.addEventListener('resize', (e) => {
    //         const width = this.canvas.clientWidth;
    //         const height = this.canvas.clientHeight;
    //         //TODO: handle resize
    //     }, false);
    // }

    bindCanvasByID = (canvasID) => {
        window.HashiraInitRenderLoop(canvasID);
    }

    sendEvent = (event, data) => {
        window.HashiraSendEvent(event, data);
    }

    loadTileset = (url, tileSize) => {
        fetch(url).then((response) => {
            return response.arrayBuffer();
        }).then((buffer) => {
            this.sendEvent("resources.LoadTileset", { tileSize: tileSize, data: new Uint8Array(buffer) });

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
        this.sendEvent("world.AddLayerData", { map: mapName, name: layerName, data: data });
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

class HashiraEditor {
    constructor(hashira) {
        this.hashira = hashira;
        this.dragStartX = 0;
        this.dragStartY = 0;
        this.dragging = false;
    }

    bindEvents = (canvas) => {
        canvas.addEventListener("mousedown", this.onCanvasMouseDown);
        canvas.addEventListener("mouseup", this.onCanvasMouseUp);
        canvas.addEventListener("mousemove", this.onCanvasMouseMove);
        canvas.addEventListener("mousemove", this.onCanvasMouseMove);
        canvas.addEventListener("contextmenu", e => e.preventDefault());
        canvas.addEventListener("wheel", this.onCanvasWheel);
    }

    onCanvasMouseDown = (e) => {
        e.preventDefault();
        if (this._isRightMB(e)) {
            this.dragStartX = e.clientX;
            this.dragStartY = e.clientY;
            this.dragging = true;
        }
    }

    onCanvasMouseUp = (e) => {
        e.preventDefault();
        this.dragging = false;
    }

    onCanvasMouseMove = (e) => {
        e.preventDefault();
        if (this.dragging) {
            const x = e.clientX;
            const y = e.clientY;

            const dx = x - this.dragStartX;
            const dy = y - this.dragStartY;

            this.dragStartX = x;
            this.dragStartY = y;

            this._rightMBDraggedBy(dx, dy);
        }
    }

    onCanvasWheel = (e) => {
        e.preventDefault();
        const by = Math.sign(e.deltaY);
        this.hashira.setCameraZoomBy(-by)
    }

    _rightMBDraggedBy = (dx, dy) => {
        this.hashira.setCameraTranslationBy(-dx, dy);
    }

    _isRightMB = (e) => {
        let isRightMB = false;
        if ("which" in e)  // Gecko (Firefox), WebKit (Safari/Chrome) & Opera
            isRightMB = e.which == 3;
        else if ("button" in e)  // IE, Opera 
            isRightMB = e.button == 2;
        return isRightMB;
    }

}