import { h, Fragment } from "preact";
import { useEffect, useRef } from "preact/hooks";
import styled from "styled-components";
import CallbackManager from "../../../helpers/CallbackManager";
import { setStickies } from "../../../helpers/oobe/stickyManager";
import { SetupOption } from "../../../types/oobe";

const StyledSelect = styled.select`
    width: 100%;
    line-height: 30px;
    padding: 5px;
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
            const res = {[props.option.id]: ref.current!.value.trim()};
            if (props.option.sticky) setStickies(res);
            return res;
        });
        return () => props.cbManager.delete(callbackId);
    }, []);

    // Get the list items.
    const items = props.option.list_items.map(
        (x, i) =>
            <option value={x.value} key={i}>{x.name}</option>);

    // Return the structure.
    return <>
        <hr />
        <h3>{props.option.name}</h3>
        <p>{props.option.description}</p>
        <StyledSelect ref={ref}>
            {items}
        </StyledSelect>
        <br /><br />
    </>;
};
