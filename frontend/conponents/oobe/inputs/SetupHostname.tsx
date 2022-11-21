import { h, Fragment } from "preact";
import CallbackManager from "../../../helpers/CallbackManager";
import { SetupOption } from "../../../types/oobe";
import { useEffect, useState } from "preact/hooks";

type Props = {
    option: SetupOption;
    cbManager: CallbackManager<{[key: string]: any}>
};

export default (props: Props) => {
    // Handle form submits.
    useEffect(() => {
        const callbackId = props.cbManager.new(() => {
            // TODO: Check the secure flag.
            return {[props.option.id]: {
                hostname: window.location.hostname,
                protocol: window.location.protocol.substring(
                    0, window.location.protocol.length - 1),
            }};
        });
        return () => props.cbManager.delete(callbackId);
    }, []);

    // Return the structure.
    return <> // TODO
        <hr />
    </>;
};
