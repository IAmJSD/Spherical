"use strict";

// Enriches the DOM element key that is specified in the data-enrich attribute with the value from the object.
// The value in the object can be an async function, a function, or anything else (not treated as callable). As a
// special case, a "!" in front of the string will be used to treat it as a boolean. Enrichment's should be split with
// a comma and use an equals to designate them.

const enrichments = {};

// Handles making sure the value gets banged enough. I love writing code sometimes.
const bangValue = (value, bangs) => {
    while (bangs > 0) {
        bangs--;
        value = !value;
    }
    return value;
};

// Enriches the DOM element key that is specified in the data-enrich attribute with the value from the object.
const enrich = node => {
    (node.dataset.enrich || "").split(",").forEach(part => {
        let [domObjKey, valueStr] = part.split("=");

        // Check the key exists on the DOM object.
        if (!domObjKey || !node[domObjKey]) {
            throw new Error(`Invalid enrichment DOM key: ${domObjKey}`);
        }

        // Check if this should do boolean logic.
        let bangs = 0;
        while (valueStr[0] === "!") {
            bangs++;
            valueStr = valueStr.slice(1);
        }

        // Check the enrichment exists.
        const enrichment = enrichments[valueStr];
        if (enrichment === undefined) {
            throw new Error(`Invalid enrichment: ${valueStr}`);
        }

        if (typeof enrichment === "function") {
            // If the enrichment is a function, call it and handle the result.
            const c = enrichment(node);
            if (c instanceof Promise) {
                c.then(value => {
                    node[domObjKey] = bangValue(value, bangs);
                });
                return;
            }
            node[domObjKey] = bangValue(c, bangs);
        } else {
            // Enrichment is not a function, just set the value.
            node[domObjKey] = bangValue(enrichment, bangs);
        }
    });
};

// Setup and start the mutation observer.
const mutationObserver = new MutationObserver(mutations => {
    for (const mutation of mutations) {
        for (const node of mutation.addedNodes) {
            if (node.nodeType === Node.ELEMENT_NODE) {
                enrich(node);
            }
        }
    }
});
mutationObserver.observe(document.body, { childList: true, subtree: true });
