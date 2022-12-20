import { decode } from "@msgpack/msgpack";

export class NonMsgpackError extends Error {
    public response: Response;

    constructor(resp: Response) {
        super("The response is not msgpack");
        this.response = resp;
    }
}

export default async (path: string) => {
    const res = await fetch(path, {
        headers: {
            Accept: "application/msgpack",
        },
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
