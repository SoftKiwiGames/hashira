"use strict";

const Hashira = {
    Init: () => {
        if (!WebAssembly.instantiateStreaming) {
            WebAssembly.instantiateStreaming = async (resp, importObject) => {
                const source = await (await resp).arrayBuffer();
                return await WebAssembly.instantiate(source, importObject);
            };
        }

        const canvases = document.getElementsByTagName("canvas");
        for (var i = 0; i < canvases.length; i++) {
            const canvas = canvases[i];
            const url = canvas.getAttribute("data-wasm-url");
            const id = canvas.getAttribute("id");

            if (!id) {
                canvas.setAttribute("id", "hashira-container-" + i);
            }
            let go = new Go();
            let mod, inst;
            WebAssembly.instantiateStreaming(fetch(url), go.importObject).then(
                (result) => {
                    go.run(result.instance)
                }
            );
        }
    },
}