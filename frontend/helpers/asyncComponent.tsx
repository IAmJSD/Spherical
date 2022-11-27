import { h, Fragment, FunctionalComponent } from "preact";
import { useEffect, useState } from "preact/hooks";

type asyncImporter<T> = () => Promise<{default: FunctionalComponent<T>}>;

export default <T,>(fn: asyncImporter<T>) => (props: T) => {
    // Handle loading the component in.
    const [val, setVal] = useState<any>();
    useEffect(() => {
        fn().then(x => setVal(x.default(props)));
    }, [props]);

    // Display nothing here until it loads in.
    if (val) return val;
    return <Fragment />;
};
