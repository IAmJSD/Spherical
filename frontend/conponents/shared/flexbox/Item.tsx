import styled from "styled-components";

type Props = {
    flexGrow?: number;
};

export default styled.div<Props>`
    flex-grow: ${props => props.flexGrow || "0"};
`;
