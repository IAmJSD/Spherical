import { h, Fragment } from "preact";
import { useEffect, useRef } from "preact/hooks";
import styled from "styled-components";
import CallbackManager from "../../../helpers/CallbackManager";
import { setStickies } from "../../../helpers/oobe/stickyManager";
import { POError, SetupOption } from "../../../types/oobe";

const StyledInput = styled.input`
    width: 100%;
    line-height: 30px;
`;

type Props = {
    option: SetupOption;
    cbManager: CallbackManager<{[key: string]: any}>;
    secret: boolean;
};

export default (props: Props) => {
    // Handle form submits.
    const ref = useRef<HTMLInputElement | undefined>();
    useEffect(() => {
        const callbackId = props.cbManager.new(() => {
            // Get the text content.
            let textContent = ref.current!.value;
            if (textContent === "") {
                // Handle blank fields.
                if (props.option.required) throw new POError(
                    "frontend/components/oobe/inputs/SetupInput:input_required");
                textContent = undefined;
            } else {
                // Handle regex validation if specified.
                if (props.option.regexp !== undefined) {
                    const regex = new RegExp(props.option.regexp);
                    if (!textContent.match(regex)) throw new POError(
                        "frontend/components/oobe/inputs/SetupInput:input_invalid");
                }
            }

            // Return the result.
            const res = {[props.option.id]: textContent};
            if (props.option.sticky) setStickies(res);
            return res;
        });
        return () => props.cbManager.delete(callbackId);
    }, []);

    // Return the structure.
    return <>
        <hr />
        <h3>{props.option.name}</h3>
        <p>{props.option.description}</p>
        <StyledInput
            type={props.secret ? "password" : "text"} ref={ref}
            placeholder={props.option.name} />
        <br /><br />
    </>;
};
