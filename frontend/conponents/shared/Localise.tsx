import { h, Fragment } from "preact";
import { useEffect, useState } from "preact/hooks";
import { decode } from "@msgpack/msgpack";
import CallbackManager from "../../helpers/CallbackManager";
import Markdown from "./Markdown";

// Start a loop up to keep trying to fetch the localisation strings. We generally
// shouldn't do stuff like this in the codebase, but this cannot fail for long
// since it severely impedes UX.
export let i18nCallbacks: CallbackManager<void> | null = new CallbackManager();
export let strings: {[key: string]: string};
const hn = () => {
    fetch("/api/internal/i18n").then(async x => {
        if (!x.ok) throw new Error(`Request returned ${x.status}`);
        strings = decode(await x.arrayBuffer()) as {[key: string]: string};
        const y = i18nCallbacks;
        i18nCallbacks = null;
        y.runAll();
        console.log("Loaded i18n", strings);
    }).catch(err => {
        console.error("Localisation get failed:", err, " - retrying in 100ms");
        setTimeout(hn, 100);
    });
};
hn();

type MarkdownOptions = {
    unsafe: boolean,
};

type Props = {
    values?: {[key: string]: any},
    i18nKey: string,
    markdownOptions?: MarkdownOptions,
};

export const lazyValues = (props: Props) =>
    props.values ? Object.keys(props.values).map(x => `{${x}}`).join(" ") : "";

export default (props: Props) => {
    // Get/set the key.
    const [key, setKey] = useState(lazyValues(props));
    useEffect(() => {
        if (i18nCallbacks) {
            const i = i18nCallbacks.new(() => setKey(strings[props.i18nKey]));
            return () => i18nCallbacks.delete(i);
        }
        setKey(strings[props.i18nKey]);
    }, []);

    // Substitute the keys.
    let res = key;
    if (props.values) {
        for (const key in props.values) {
            res = res.replace("{" + key + "}", props.values[key]);
        }
    }
    console.log(res);

    // If the markdown options does not exist, return it as a fragment here.
    if (!props.markdownOptions) return <>{res}</>;

    // Return the options as markdown.
    return <Markdown content={res} {...props.markdownOptions} />;
};

export const reload = () => fetch("/api/internal/i18n").then(async x => {
    if (!x.ok) throw new Error(`Request returned ${x.status}`);
    strings = decode(await x.arrayBuffer()) as {[key: string]: string};
    console.log("Reloaded i18n", strings);
});
