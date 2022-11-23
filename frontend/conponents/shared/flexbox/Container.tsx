import styled from "styled-components";

type Props = {
    flexDirection: "row" | "row-reverse" | "column" | "column-reverse",
    flexWrap?: boolean,
    justifyContent?: "flex-start" | "flex-end" | "center",
};

export default styled.div<Props>`
    display: flex;
    flex-direction: ${props => props.flexDirection};
    flex-wrap: ${props => props.flexWrap ? "wrap" : "nowrap"};
    ${props => props.justifyContent ? `justify-content: ${props.justifyContent};` : ""}
`;
