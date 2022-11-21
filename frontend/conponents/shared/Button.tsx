import styled from "styled-components";
import { h, ComponentChildren } from "preact";
import { useState } from "preact/hooks";
import { HTMLAttributes } from "preact/compat";

type ButtonType = "normal" | "danger";
type PropState = "clickable" | "disabled";

type AProps = {
    type: ButtonType,
    state: PropState | "loading",
};

const cursor = (props: AProps) => {
    switch (props.state) {
        case "disabled":
            return "not-allowed";
        case "clickable":
            return "default";
        case "loading":
            return "wait";
    }
};

const bg = (props: AProps) => {
    switch (props.type) {
        case "normal":
            return "linear-gradient(to right top, #196eee, #007ce5, #0085d4, #008abf, #1f8daa)";
        case "danger":
            return "linear-gradient(to right top, #963817, #9b3318, #a02e1a, #a5271c, #aa1f1f)";
    }
};

const A = styled.a<AProps>`
    padding: 10px;
    color: white;
    border-radius: 6%;
    cursor: ${cursor};
    background: ${bg};

    ${
        props => props.state === "clickable" ?
            ":hover { opacity: 0.8; }" : "opacity: 0.3;"
    }
`;

type Props = {
    submit: () => Promise<void>,
    type: ButtonType,
    state: PropState,
    children: ComponentChildren,
};

export default (props: Props) => {
    const [state, setState] = useState<PropState | "loading">(props.state);
    const submit = () => {
        setState("loading");
        props.submit().finally(() => setState("clickable"));
    };
    const additional: HTMLAttributes<HTMLLinkElement> = {};
    if (state === "clickable") additional.onClick = submit;
    return <A state={state} type={props.type} {...additional}>{props.children}</A>;
};
