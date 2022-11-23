import { h } from "preact";
import { useEffect, useState } from "preact/hooks";
import { i18nCallbacks, strings, lazyValues } from "./Localise";

type Props = {
    values?: {[key: string]: any},
    i18nKey: string,
    imageUrl: string,
};

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

    // Return the image.
    return <img src={props.imageUrl} alt={res} />;
};
