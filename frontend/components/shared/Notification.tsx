import styled from "styled-components";

type Props = {
    type: "normal" | "danger" | "warning" | "success",
    centered: boolean,
};

const bg = (props: Props) => {
    switch (props.type) {
        case "normal":
            return "linear-gradient(to right top, #196eee, #007ce5, #0085d4, #008abf, #1f8daa)";
        case "danger":
            return "linear-gradient(to right top, #963817, #9b3318, #a02e1a, #a5271c, #aa1f1f)";
        case "warning":
            return "linear-gradient(to right top, #978820, #9c831e, #a17e1d, #a6781d, #aa731f)";
        case "success":
            return "linear-gradient(to right top, #639720, #638b22, #627f25, #607427, #5c692a)";
    }
};

export default styled.p<Props>`
    background: ${bg};
    padding: 22px;
    color: white;
    ${props => props.centered ? "width: 50%; margin: auto;" : ""}
    border-radius: 2%;
`;
