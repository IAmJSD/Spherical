import { h, Fragment } from "preact";
import CallbackManager from "../../../helpers/CallbackManager";
import { SetupOption } from "../../../types/oobe";
import { useEffect, useState } from "preact/hooks";

type Props = {
    option: SetupOption;
    cbManager: CallbackManager<{[key: string]: any}>;
};

export default (props: Props) => {
    // Handle form submits.
    useEffect(() => {
        const callbackId = props.cbManager.new(() => {
            // TODO: Validate the input if un-optional.
            // TODO: Get text content.
            const textContent = "";
            return {[props.option.id]: textContent};
        });
        return () => props.cbManager.delete(callbackId);
    }, []);

    // Return the structure.
    return <>
        <hr />
        TODO
    </>;
};
