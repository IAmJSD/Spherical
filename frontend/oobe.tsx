import "preact/devtools";

import { h, render } from "preact";
import { useState, useEffect } from "preact/hooks";
import { POError, SetupOption, SetupType } from "./types/oobe";
import fetch from "./helpers/redirectingFetch";
import CallbackManager from "./helpers/CallbackManager";
import { loadStickies } from "./helpers/oobe/stickyManager";
import Layout from "./components/oobe/Layout";
import Loading from "./components/shared/Loading";
import SetupBoolean from "./components/oobe/inputs/SetupBoolean";
import SetupHostname from "./components/oobe/inputs/SetupHostname";
import SetupInput from "./components/oobe/inputs/SetupInput";
import SetupTextbox from "./components/oobe/inputs/SetupTextbox";
import Button from "./components/shared/Button";
import Localise from "./components/shared/Localise";
import Notification from "./components/shared/Notification";
import LocalisedAltText from "./components/shared/LocalisedAltText";
import asyncComponent from "./helpers/asyncComponent";

const Markdown = asyncComponent(() => import("./components/shared/Markdown"));

type InstallData = {
    image_url: string;
    image_alt: string;
    title: string;
    description: string;
    step: string;
    options: SetupOption[];
    next_button: string;
};

const validatorCallbacks: CallbackManager<{[key: string]: any}> = new CallbackManager();

const Main = () => {
    // Defines getting/defining the install data.
    const [installData, setInstallData] = useState<InstallData | null | undefined>(undefined);
    useEffect(() => {
        fetch("/install/state").then(async x => {
            if (!x.ok) throw new Error(`Initial GET got status ${x.status}`);
            const j = await x.json();
            setInstallData(j);
        }).catch(() => setInstallData(null));
    }, []);

    // Defines the error message.
    const [error, setError] = useState<string | POError>("");

    // Handle the submit button.
    const submit = () => {
        // Validate the content.
        let values: { [key: string]: any }[];
        try {
            values = validatorCallbacks.runAll();
        } catch (e) {
            if (e instanceof POError) setError(e);
            else setError(e.message);
            return;
        }

        // Make a object.
        let o: any = {};
        for (const val of values) {
            for (const key of Object.keys(val)) o[key] = val[key];
        }
        o = Object.assign(loadStickies(), o);

        // Make the fetch request.
        return fetch("/install/state", {
            method: "POST",
            body: JSON.stringify({
                type: installData.step,
                body: o,
            }),
            headers: {
                "Content-Type": "application/json",
            },
        }).then(async x => {
            if (x.status === 400) {
                // Set the message to this.
                const j = await x.json();
                setError(j.message);
                return;
            }

            if (x.ok) {
                // Set the error message blank.
                setError("");
                setInstallData(await x.json());
                return;
            }

            // Throw an error.
            throw new Error(`POST status returned ${x.status}.`);
        }).catch(() => setInstallData(null));
    };

    // If undefined, we should just show loading.
    if (installData === undefined) return <Loading />;

    // If null, return a error.
    if (installData === null) return <Layout>
        <p className="centrist">
            <LocalisedAltText
                i18nKey="frontend/oobe:cross_image"
                imageUrl="/png/cross.png"
            />
        </p>
        <h1><Localise i18nKey="frontend/oobe:failed_title" /></h1>
        <p className="centrist">
            <Localise i18nKey="frontend/oobe:failed_description" />
        </p>
    </Layout>;

    // Map the objects.
    const components = installData.options.map((x, i) => {
        switch (x.type) {
            case SetupType.BOOLEAN:
                return <SetupBoolean
                    key={`${i}_${x.id}`} option={x} cbManager={validatorCallbacks}
                />;
            case SetupType.HOSTNAME:
                return <SetupHostname
                    key={`${i}_${x.id}`} option={x} cbManager={validatorCallbacks}
                />;
            case SetupType.INPUT:
                return <SetupInput
                    key={`${i}_${x.id}`} option={x} cbManager={validatorCallbacks}
                    type="text"
                />;
            case SetupType.SECRET:
                return <SetupInput
                    key={`${i}_${x.id}`} option={x} cbManager={validatorCallbacks}
                    type="password"
                />;
            case SetupType.NUMBER:
                return <SetupInput
                    key={`${i}_${x.id}`} option={x} cbManager={validatorCallbacks}
                    type="number"
                />;
            case SetupType.TEXTBOX:
                return <SetupTextbox key={`${i}_${x.id}`} option={x} cbManager={validatorCallbacks} />;
            default:
                throw new Error("Input type not implemented.");
        }
    });

    // Figure out the component to show in the error box.
    const errorChild = error instanceof POError ? <Localise i18nKey={error.message} /> : error;

    // Return the layout.
    return <Layout>
        <p className="centrist">
            <img src={installData.image_url} alt={installData.image_alt} />
        </p>
        <h1>{installData.title}</h1>
        <Markdown content={installData.description} unsafe={true} />
        {components}
        <hr />
        {
            error === "" ? null : <div className="centrist" style={{marginTop: "30px"}}>
                <Notification type="danger" centered={true}>
                    {errorChild}
                </Notification>
            </div>
        }
        <div className="centrist" style={{marginTop: "30px"}}>
            <p>
                <Button submit={submit} type="normal" state="clickable">
                    {installData.next_button}
                </Button>
            </p>
        </div>
    </Layout>;
};

render(<Main />, document.getElementById("app_mount"));
