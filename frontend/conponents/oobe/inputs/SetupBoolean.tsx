import { h, Fragment } from "preact";
import { useEffect, useRef } from "preact/hooks";
import styled from "styled-components";
import CallbackManager from "../../../helpers/CallbackManager";
import { SetupOption } from "../../../types/oobe";
import Container from "../../shared/flexbox/Container";
import Item from "../../shared/flexbox/Item";

type Props = {
    option: SetupOption;
    cbManager: CallbackManager<{[key: string]: any}>
};

const BigCheck = styled.input`
    transform: scale(2);
`;

export default (props: Props) => {
    // Handle form submits.
    const ref = useRef<HTMLInputElement | undefined>();
    useEffect(() => {
        const callbackId = props.cbManager.new(() => {
            return {[props.option.id]: ref.current!.checked};
        });
        return () => props.cbManager.delete(callbackId);
    }, []);

    // Return the structure.
    return <>
        <hr />
        <Container flexDirection="row">
            <Item>
                <div style={{padding: "25px"}}>
                    <BigCheck type="checkbox" ref={ref} />
                </div>
            </Item>
            <Item>
                <h3>{props.option.name}</h3>
                <p>{props.option.description}</p>
            </Item>
        </Container>
    </>;
};
