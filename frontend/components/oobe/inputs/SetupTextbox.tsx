import { h, Fragment } from "preact";
import { useEffect, useRef } from "preact/hooks";
import styled from "styled-components";
import CallbackManager from "../../../helpers/CallbackManager";
import { setStickies } from "../../../helpers/oobe/stickyManager";
import { POError, SetupOption } from "../../../types/oobe";

const StyledTextarea = styled.textarea`
    width: 100%;
    line-height: 30px;
    resize: vertical;
`;

type Props = {
    option: SetupOption;
    cbManager: CallbackManager<{[key: string]: any}>;
};

export default (props: Props) => {
    // Handle form submits.
    const ref = useRef<HTMLInputElement | undefined>();
    useEffect(() => {
        const callbackId = props.cbManager.new(() => {
            // Get the text content.
            let textContent: string | number | undefined = ref.current!.value.trim();
            if (textContent === "") {
                // Handle blank fields.
                if (props.option.required) throw new POError(
                    "frontend/components/oobe/inputs/SetupTextbox:input_required");
                textContent = undefined;
            } else {
                // Handle regex validation if specified.
                if (props.option.regexp !== undefined) {
                    const regex = new RegExp(props.option.regexp);
                    if (!textContent.match(regex)) throw new POError(
                        "frontend/components/oobe/inputs/SetupTextbox:input_invalid");
                }
            }

            // Return the result.
            if (props.option.type === "number") textContent = Number(textContent);
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
        <StyledTextarea
            ref={ref}
            placeholder={props.option.name} />
        <br /><br />
    </>;
};
