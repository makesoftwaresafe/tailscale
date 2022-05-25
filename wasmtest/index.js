import "./wasm_exec";
import "./notifier";
import "./js-state-store"

const go = new window.Go();
WebAssembly.instantiateStreaming(fetch("test.wasm"), go.importObject).then((result) => {
    go.run(result.instance);
});
