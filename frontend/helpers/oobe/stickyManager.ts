export const loadStickies = (): {[key: string]: string} => {
    const s = localStorage.getItem("setup_stickies");
    if (s) return JSON.parse(s);
    return {};
};

export const setStickies = (result: {[ley: string]: any}): {[key: string]: any} => {
    const s = loadStickies();
    Object.assign(s, result);
    localStorage.setItem("setup_stickies", JSON.stringify(s));
    return result;
};
