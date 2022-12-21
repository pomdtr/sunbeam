// See the Electron documentation for details on how to use preload scripts:
// https://www.electronjs.org/docs/latest/tutorial/process-model#preload-scripts

import { contextBridge, ipcRenderer } from "electron";

contextBridge.exposeInMainWorld("electron", {
  showWindow: () => ipcRenderer.invoke("showWindow"),
  hideWindow: () => ipcRenderer.invoke("hideWindow"),
  copyToClipboard: (text: string) =>
    ipcRenderer.invoke("copyToClipboard", text),
  openInBrowser: (url: string) => ipcRenderer.invoke("openInBrowser", url),
});
