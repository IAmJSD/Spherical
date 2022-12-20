import { decode } from "@msgpack/msgpack";

export class NonMsgpackError extends Error {
    public response: Response;

    constructor(resp: Response) {
        super("The response is not msgpack");
        this.response = resp;
    }
}

export default async (path: string, method: string, body: any) => {
    const j = JSON.stringify(body);
    const res = await fetch(path, {
        headers: {
            "Content-Type": "application/json",
            "Content-Length": j.length.toString(),
            Accept: "application/msgpack",
        },
        body: j,
        method: method,
    });
    if (
        !["application/msgpack", "application/x-msgpack"].
        includes(res.headers.get("Content-Type").toLowerCase())
    ) {
        throw new NonMsgpackError(res);
    }
    return {
        body: decode(await res.arrayBuffer()),
        headers: res.headers,
        status: res.status,
        ok: res.ok,
    };
};
