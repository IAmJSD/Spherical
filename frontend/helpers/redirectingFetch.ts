const redirectingFetch: typeof fetch = (input, init) => {
    if (!init) init = {};
    init.redirect = "follow";
    return fetch(input, init).then(x => {
        if (x.redirected) {
            // Set the window location to where we want to go.
            window.location.replace(x.url);
            return new Promise(() => {});
        }
        return x;
    });
};

export default redirectingFetch;
