import { h } from "preact";
import { useEffect, useRef, useState } from "preact/hooks";
import styled from "styled-components";

let num = 0;

const InputContainer = styled.div`
    text-align: left;
    margin: 1%;
`;

const Input = styled.input`
    width: 99%;
    line-height: 25px;
    margin-top: 5px;
    margin-bottom: 5px;
    border-radius: 8px;
`;

type Props = {
    children: any;
    trimSpaces: boolean;
    type: "password" | "text" | "number";
    onSubmit: (hn: () => string) => void;
};

export default ({ children, trimSpaces, type, onSubmit }: Props) => {
    const [id] = useState(`id__${num++}`);
    const [placeholder, setPlaceholder] = useState("");
    const labelRef = useRef<HTMLLabelElement | undefined>();
    const inputRef = useRef<HTMLInputElement | undefined>();

    useEffect(() => {
        onSubmit(() => {
            let text = inputRef.current!.value;
            if (trimSpaces) text = text.trim();
            return text;
        });
    }, []);

    // Put the text contents of the label into placeholder.
    if (labelRef.current) {
        setPlaceholder(labelRef.current.textContent);
    }

    return <InputContainer>
        <label for={id} ref={labelRef} style={{marginLeft: "2px"}}>
            {children}:
        </label>
        <Input id={id} type={type} ref={inputRef} placeholder={placeholder} />
    </InputContainer>;
};
