import styled from "styled-components";

type Props = {
    type: "normal" | "danger",
};

const bg = (props: Props) => {
    switch (props.type) {
        case "normal":
            return "linear-gradient(to right top, #196eee, #007ce5, #0085d4, #008abf, #1f8daa)";
        case "danger":
            return "linear-gradient(to right top, #963817, #9b3318, #a02e1a, #a5271c, #aa1f1f)";
    }
};

export default styled.p<Props>`
    background: ${bg};
    padding: 22px;
    color: white;
    width: 50%;
    margin: auto;
    border-radius: 2%;
`;
