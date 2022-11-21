import { h, Fragment } from "preact";
import CallbackManager from "../../../helpers/CallbackManager";
import { SetupOption } from "../../../types/oobe";
import { useEffect, useState } from "preact/hooks";

type Props = {
    option: SetupOption;
    cbManager: CallbackManager<{[key: string]: any}>
};

export default (props: Props) => {
    // Defines the boolean value.
    const [checked, setChecked] = useState(false);

    // Handle form submits.
    useEffect(() => {
        const callbackId = props.cbManager.new(() => {
            return {[props.option.id]: checked};
        });
        return () => props.cbManager.delete(callbackId);
    }, []);

    // Return the structure.
    return <> // TODO
        <hr />
    </>;
};
