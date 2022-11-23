import { h, Fragment } from "preact";
import { useEffect } from "preact/hooks";
import CallbackManager from "../../../helpers/CallbackManager";
import { POError, SetupOption } from "../../../types/oobe";
import Notification from "../../shared/Notification";
import Localise from "../../shared/Localise";

type Props = {
    option: SetupOption;
    cbManager: CallbackManager<{[key: string]: any}>
};

export default (props: Props) => {
    // Defines the security state. Is false/true if allowed to go ahead, or null if not.
    let securityState: boolean | null = window.location.protocol === "https:";
    if (!securityState) {
        // Check if we are validating security and if so null it.
        if (props.option.must_secure) securityState = null;
    }

    // Handle form submits.
    useEffect(() => {
        const callbackId = props.cbManager.new(() => {
            if (securityState === null) throw new POError(
                "frontend/components/oobe/inputs/SetupHostname:please_https");
            return {[props.option.id]: {
                hostname: window.location.hostname,
                protocol: window.location.protocol.substring(
                    0, window.location.protocol.length - 1),
            }};
        });
        return () => props.cbManager.delete(callbackId);
    }, []);

    // Get the i18m key.
    let i18nKey = "frontend/components/oobe/inputs/SetupHostname:ssl_okay";
    switch (securityState) {
        case null:
            i18nKey = "frontend/components/oobe/inputs/SetupHostname:ssl_error";
            break;
        case false:
            i18nKey = "frontend/components/oobe/inputs/SetupHostname:ssl_warning";
            break;
    }

    // Return the structure.
    const notifType = securityState === null ? "danger" : securityState ? "success" : "warning";
    return <>
        <hr />
        <h3>{props.option.name}</h3>
        <Notification type={notifType} centered={false}>
            <Localise i18nKey={i18nKey} />
        </Notification>
    </>;
};
