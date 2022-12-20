import styled from "styled-components";

type Props = {
    backgroundUrl: string;
};

export default styled.div<Props>`
    ${props => {
        // Set the background image.
        if (props.backgroundUrl) {
            return `background-image: url(${props.backgroundUrl});`;
        }

        // Set a grey background.
        return `background-color: #f5f5f5;
                background-image: linear-gradient(135deg, #f5f5f5 25%, transparent 25%, transparent 50%, #f5f5f5 50%, #f5f5f5 75%, transparent 75%, transparent);`;
    }}
    background-size: cover;
    background-position: center;
    background-repeat: no-repeat;
    position: fixed;
    filter: blur(8px);
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: -1;
`;
