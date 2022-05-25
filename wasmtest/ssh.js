import { Terminal } from 'xterm';
import { WebglAddon } from 'xterm-addon-webgl';

export function ssh(hostname) {
  const termContainerNode = document.createElement("div");
  termContainerNode.classname = "term-container";
  document.body.appendChild(termContainerNode);

  const term = new Terminal({
    cursorBlink: true
  });
  term.open(termContainerNode);

  try {
    term.loadAddon(new WebglAddon());
  } catch (e) {
    console.warn('WebGL addon threw an exception during load', e);
  }

  // Cancel wheel events from scrolling the page if the terminal has scrollback
  termContainerNode.addEventListener('wheel', e => {
    if (term.buffer.active.baseY > 0) {
      e.preventDefault();
    }
  });

  let onDataHook;
  term.onData(e => {
    onDataHook?.(e);
  });

  term.focus();

  runSSH(
    hostname,
    input => term.write(input),
    hook => onDataHook = hook,
    term.rows,
    term.cols,
    () => {
      term.dispose();
      termContainerNode.remove();
    });

}
