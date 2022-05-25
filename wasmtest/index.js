import "./wasm_exec";
import "./term";

const go = new window.Go();
WebAssembly.instantiateStreaming(fetch("test.wasm"), go.importObject).then((result) => {
    go.run(result.instance);
});
