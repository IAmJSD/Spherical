export default class CallbackManager<T> {
    private m: Map<number, (() => T)>;
    private i: number;

    constructor() {
        this.m = new Map();
        this.i = 0;
    }

    new(cb: () => T): number {
        const i = this.i;
        this.i++;
        this.m.set(i, cb);
        return i;
    }

    delete(i: number) {
        this.m.delete(i);
    }

    runAll(): T[] {
        return [...this.m.values()].map(x => x());
    }
}
