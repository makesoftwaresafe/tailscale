import QRCode from "qrcode";

/**
 * @fileoverview Notification callback functions (bridged from ipn.Notify)
 */

const State = {
    NoState: 0,
    InUseOtherUser: 1,
    NeedsLogin: 2,
    NeedsMachineAuth: 3,
    Stopped: 4,
    Starting: 5,
    Running: 6,
};

 globalThis.notifyState = function(state) {
    const stateNode = document.getElementById("state");
    stateNode.textContent = `Unknown (${state})`;
    for (const [name, value] of Object.entries(State)) {
        if (state === value) {
            stateNode.textContent = name;
            break;
        }
    }
    if (state === State.Running || state === State.Starting) {
        const loginNode = document.getElementById("loginURL");
        loginNode.innerHTML = "";
    }
 };

 globalThis.notifyNetMap = function(netMapStr) {
    const netMap = JSON.parse(netMapStr);
    const netmapMode = document.getElementById("netMap");
    netmapMode.textContent = JSON.stringify(netMap, null, 2);
  };

  globalThis.notifyBrowseToURL = async function(url) {
    const loginNode = document.getElementById("loginURL");
    loginNode.innerHTML = "";
    const linkNode = document.createElement("a");
    linkNode.href = url;
    linkNode.target = "_blank";
    linkNode.textContent = url;
    loginNode.appendChild(linkNode);

    try {
      const dataURL = await QRCode.toDataURL(url, {width: 512});
      linkNode.appendChild(document.createElement("br"));
      const imageNode = document.createElement("img");
      imageNode.src = dataURL;
      imageNode.width = 256;
      imageNode.height = 256;
      imageNode.border = "0";
      linkNode.appendChild(imageNode);
    } catch (err) {
      console.error("Could not generate QR code:", err);
    }
  };
