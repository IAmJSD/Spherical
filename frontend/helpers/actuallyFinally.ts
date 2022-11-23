// Handles errors before the promise as well as the .finally within the promise.

export default <T>(fn: () => Promise<T>, finallyHn: () => void) => {
    let p;
    try {
        p = fn();
    } catch (_) {
        finallyHn();
        return;
    }
    return p ? p.finally(finallyHn) : finallyHn();
};
