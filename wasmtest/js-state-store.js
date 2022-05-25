/**
 * @fileoverview Callbacks used by jsStateStore to persist IPN state.
 */

globalThis.setState = (id, value) => window.sessionStorage[`ipn-state-${id}`] = value;
globalThis.getState = (id) => window.sessionStorage[`ipn-state-${id}`] || "";
