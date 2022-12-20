import { PayloadID, WebSocketPayload, parse, serialize } from "../types/payloads";

// Defines the no reconnect handlers.
const noReconnectHandlers = new Map<number, () => void>();

// Defines the reconnect handlers.
const reconnectHandlers = new Map<number, () => void>();

// Defines the next ID.
let nextId = 0;

// Adds a disconnect handler.
export const addDisconnectHandler = (hn: (reconnect: boolean) => void) => {
    const id = nextId++;
    noReconnectHandlers.set(id, () => hn(false));
    reconnectHandlers.set(id, () => hn(true));
    return id;
};

// Removes a disconnect handler.
export const removeDisconnectHandler = (id: number) => {
    noReconnectHandlers.delete(id);
    reconnectHandlers.delete(id);
};

// Defines the ready handlers.
const readyHandlers = new Map<number, () => void>();

// Adds a ready handler.
export const addReadyHandler = (hn: () => void) => {
    const id = nextId++;
    readyHandlers.set(id, hn);
    return id;
};

// Removes a ready handler.
export const removeReadyHandler = (id: number) => {
    readyHandlers.delete(id);
};

// Defines the websocket.
let ws: WebSocket | null = null;

// Defines the websocket stage.
export let wsStage: "negotiating" | "ready" | "unhealthy" | "closed" = "closed";

// Handles a do not reconnect case.
const noReconnect = () => {
    wsStage = "unhealthy";
    noReconnectHandlers.forEach(hn => hn());
};

// Defines the timeout for inbound.
let inboundTimeout: number | null = null;

// Defines the heartbeat interval.
let heartbeatInterval: number | null = null;

// Handle the websocket closing.
const wsClose = (ev: CloseEvent) => {
    let body = {
        reason: "unable to determine reason from payload",
        reconnect: false,
    };
    try {
        body = JSON.parse(ev ? ev.reason : "");
    } catch (_) {}
    console.error("websocket closed", body);
    if (inboundTimeout !== null) clearTimeout(inboundTimeout);
    if (body.reconnect) {
        // Set ws to null and call the start function.
        ws = null;
        wsStage = "unhealthy";
        startWebsocket();
        reconnectHandlers.forEach(hn => hn());
    } else {
        // Handle the no reconnect case.
        noReconnect();
    }
};

// Sends a payload.
export const sendPayload = (payload: WebSocketPayload) => {
    if (ws === null) throw new Error("websocket not connected");
    ws.send(serialize(payload));
};

// Defines the message handler.
// TODO: zlib
const wsMessage = (ev: MessageEvent) => {
    const data = ev.data as ArrayBuffer;
    const payload = parse(new Uint8Array(data));
    switch (payload.id) {
        case PayloadID.ACCEPTED_PAYLOAD: {
            console.log("got accepted payload!", payload);
            heartbeatInterval = payload.body.heartbeat_interval;
            break;
        }
        case PayloadID.HEARTBEAT_PAYLOAD: {
            sendPayload(payload);
            break;
        }
        case PayloadID.READY_PAYLOAD: {
            console.log("got ready payload!", payload);
            wsStage = "ready";
            // TODO: handle guild events
            readyHandlers.forEach(hn => hn());
            break;
        }
    }
};

// Starts the websocket.
// TODO: zlib
export const startWebsocket = () => {
    // Return if the websocket is already present.
    if (ws) return;

    // Send a websocket to ws(s)://<host>/api/v1/gateway.
    ws = new WebSocket(`ws${location.protocol === "https:" ? "s" : ""}://${location.host}/api/v1/gateway`);

    // Set the heartbeat interval to null.
    heartbeatInterval = null;

    // Set the stage to negotiating.
    wsStage = "negotiating";

    // Set the binary type to array buffer.
    ws.binaryType = "arraybuffer";

    // Set the websocket close handler.
    ws.onclose = wsClose;
    ws.onerror = err => {
        console.error("websocket error", err);
        if (inboundTimeout !== null) clearTimeout(inboundTimeout);
        ws = null;
        wsStage = "unhealthy";
        noReconnectHandlers.forEach(hn => hn());
        startWebsocket();
    };

    // Set the websocket message handler.
    ws.onmessage = wsMessage;

    // Send the hello payload.
    ws.onopen = () => {
        sendPayload({
            id: PayloadID.HELLO_PAYLOAD,
            body: {
                token: localStorage.getItem("token"),
                cross_node: false,
            },
        });
    };
};
