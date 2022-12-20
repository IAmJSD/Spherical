import { encode, decode } from "@msgpack/msgpack";

export enum PayloadID {
    HELLO_PAYLOAD,
    ACCEPTED_PAYLOAD,
    HEARTBEAT_PAYLOAD,
    JOIN_GUILD_PAYLOAD,
    READY_PAYLOAD,
    GUILD_UPDATE_PAYLOAD,
}

export type HelloPayload = {
    id: PayloadID.HELLO_PAYLOAD,
    body: {
        token: string,
        cross_node: false,
    },
};

export type AcceptedPayload = {
    id: PayloadID.ACCEPTED_PAYLOAD,
    body: {
        heartbeat_interval: number,
    },
};

export type HeartbeatPayload = {
    id: PayloadID.HEARTBEAT_PAYLOAD,
    body: {
        id: string,
    },
};

export type JoinGuildPayload = {
    id: PayloadID.JOIN_GUILD_PAYLOAD,
    body: {
        hostname: string,
        invite_code: string,
        reply_id: string,
    },
};

export type ReadyPayload = {
    id: PayloadID.READY_PAYLOAD,
    body: {
        // TODO
    },
};

export type GuildUpdatePayload = {
    id: PayloadID.GUILD_UPDATE_PAYLOAD,
    body: {
        // TODO
    },
};

export type WebSocketPayload =
    HelloPayload | AcceptedPayload | HeartbeatPayload |
    JoinGuildPayload | ReadyPayload | GuildUpdatePayload;

export const parse = (data: Uint8Array): WebSocketPayload => {
    const id = data[0] << 8 | data[1];
    const body = decode(data.slice(2)) as any;
    return {id, body};
};

export const serialize = (payload: WebSocketPayload): Uint8Array => {
    const body = encode(payload.body);
    const data = new Uint8Array(2 + body.length);
    data[0] = payload.id >> 8;
    data[1] = payload.id & 0xff;
    data.set(body, 2);
    return data;
};
