import { h, Fragment } from "preact";
import { Navigate, useLocation } from "react-router-dom";
import { useState, useEffect, useRef } from "preact/hooks";
import styled from "styled-components";
import msgpackGet from "../../../helpers/msgpackGet";
import msgpackBody from "../../../helpers/msgpackBody";
import Background from "../Background";
import Container from "../../shared/flexbox/Container";
import Item from "../../shared/flexbox/Item";
import asyncComponent from "../../../helpers/asyncComponent";
import Localise from "../../shared/Localise";
import Input from "../Input";
import Button from "../../shared/Button";
import Notification from "../../shared/Notification";

const Markdown = asyncComponent(() => import("../../shared/Markdown"));

type NodeInfo = {
    server_name: string;
    server_description: string;
    server_background: string;
    locale: string;
    public: boolean;
    sign_ups_enabled: boolean;
};

type DescriptionProps = {
    nodeInfo: NodeInfo | null | undefined;
};

const BoxWrapper = styled.div`
    background-color: #3d3b3b;
    height: 100%;
    overflow: auto;
    border-radius: 20px;
    box-shadow: 0 0 10px 0 rgba(0, 0, 0, 0.5);
`;

const DescriptionWrapper = styled.div`
    padding: 3%;
    padding-top: 2%;
    color: #fff;
    font-family: 'Roboto Regular', Arial, Helvetica, sans-serif;
`;

const Description = ({ nodeInfo }: DescriptionProps) => {
    // Return a empty fragment if the node info is not loaded yet.
    if (!nodeInfo) return <></>;

    // Return the content.
    return <BoxWrapper>
        <DescriptionWrapper>
            <Markdown unsafe={true} content={nodeInfo.server_description} />
        </DescriptionWrapper>
    </BoxWrapper>;
};

type AuthenticationProps = {
    setU2fStage: (stage: boolean) => void;
    setToken: (token: string) => void;
    nodeInfo: NodeInfo | null | undefined;
};

enum MFAMethod {
    TOTP = "totp",
    RECOVERY = "recovery",
}

type MFAResponse = {
    supported_methods: MFAMethod[];
    half_token: string;
};

type MFAProps = {
    mfaResponse: MFAResponse;
    setMfaResponse: (mfaResponse: MFAResponse | null) => void;
    setToken: (token: string) => void;
};

const MFA = ({ mfaResponse, setMfaResponse, setToken }: MFAProps) => {
    // TODO
    return <span />;
};

type LoginFormProps = {
    setToken: (token: string) => void;
    setMfaResponse: (mfaResponse: MFAResponse | null) => void;
    nodeInfo: NodeInfo | null | undefined;
};

const LoginContainer = styled.div`
    width: 100%;
    display: block;
    text-align: center;
    margin: auto;
    font-family: 'Roboto Regular', Arial, Helvetica, sans-serif;
    color: #fff;
`;

const LoginForm = ({ setToken, setMfaResponse, nodeInfo }: LoginFormProps) => {
    // Defines the handler.
    const usernameCbRef = useRef<(() => string) | null>(null);
    const passwordCbRef = useRef<(() => string) | null>(null);
    const [error, setError] = useState<string | {po: string} | null>(null);
    const submit = async () => {
        // Get the username and password.
        const username = usernameCbRef.current!();
        const password = passwordCbRef.current!();

        // Check if the username or password is empty.
        if (username === "" || password === "") {
            setError({
                po: "frontend/components/main/routes/Root:empty",
            });
        }

        // Set the error to null whilst we wait.
        setError(null);

        // Make the request.
        const res = await msgpackBody("/api/v1/auth/password", "POST", {
            username, password,
        });
        if (res.ok) {
            // Set the token.
            const body = res.body as {token: string};
            setToken(body.token);
            return;
        }

        const code = (res.body as {code: string}).code;
        if (code === "half_authenticated") {
            // Set the mfa response.
            const body = (res.body as {body: MFAResponse}).body;
            setMfaResponse(body);
            return;
        }

        const body = (res.body as {body: any}).body;
        setError("message" in body ? body.message : body);
    };

    // Return the main form.
    return <LoginContainer>
        <h1><Localise
            values={{nodeName: nodeInfo ? nodeInfo.server_name : ""}}
            i18nKey="frontend/components/main/routes/Root:welcome"
        /></h1>
        <p>
            <Localise
                i18nKey="frontend/components/main/routes/Root:enter_username_and_password"
            />
        </p>
        {
            error ? <Notification type="danger" centered={true}>
                {
                    typeof error === "string" ? error : <Localise
                        i18nKey={error.po}
                    />
                }
            </Notification> : null
        }
        <Input trimSpaces={true} type="text" onSubmit={x => usernameCbRef.current = x}>
            <Localise
                i18nKey="frontend/components/main/routes/Root:username_or_email_address"
            />
        </Input>
        <Input trimSpaces={true} type="password" onSubmit={x => passwordCbRef.current = x}>
            <Localise
                i18nKey="frontend/components/main/routes/Root:password"
            />
        </Input>
        <p>
            <Button submit={submit} type="normal" state="clickable">
                <Localise
                    i18nKey="frontend/components/main/routes/Root:login"
                />
            </Button>
        </p>
    </LoginContainer>;
};

const Authentication = ({ setU2fStage, setToken, nodeInfo }: AuthenticationProps) => {
    const [mfaResponse, setMfaResponse] = useState<MFAResponse | null>(null);
    const setter = (body: MFAResponse | null) => {
        setU2fStage(!!body);
        setMfaResponse(body);
    };
    return <BoxWrapper>
        {
            mfaResponse ? <MFA mfaResponse={mfaResponse} setMfaResponse={setter} setToken={setToken} /> :
                <LoginForm setMfaResponse={setter} setToken={setToken} nodeInfo={nodeInfo} />
        }
    </BoxWrapper>;
};

type LoginProps = {
    setToken: (token: string) => void;
    nodeInfo: NodeInfo | null | undefined;
};

const RootContainer = styled.div`
    width: 75%;
    height: 100%;
    min-height: 100%;
    margin: auto;
    z-index: 100;
`;

const Login = ({ setToken, nodeInfo }: LoginProps) => {
    const [u2fStage, setU2fStage] = useState(false);
    return <>
        <Background backgroundUrl={nodeInfo ? nodeInfo.server_background : ""} />
        <RootContainer>
            <Container flexDirection="row" flexWrap={true} style={{height: "80%", marginTop: "4%"}}>
                <Item flexGrow={59}>
                    <Authentication
                        setU2fStage={setU2fStage} setToken={setToken}
                        nodeInfo={nodeInfo} />
                </Item>
                {
                    u2fStage ? undefined : <>
                        <Item flexGrow={2}></Item>
                        <Item flexGrow={39}>
                            <Description nodeInfo={nodeInfo} />
                        </Item>
                    </>
                }
            </Container>
        </RootContainer>
    </>;
};

export default () => {
    // Handle the fetch request if we shouldn't just be refreshing anyway.
    const location = useLocation();
    const [nodeInfo, setNodeInfo] = useState<NodeInfo | null | undefined>(undefined);
    const [hasToken, setHasToken] = useState(!!localStorage.getItem("token"));
    useEffect(() => {
        if (hasToken) {
            // Token set. Just return here.
            return;
        }

        // Do the network request.
        msgpackGet("/api/v1/node").then(x => {
            if (!x.ok) throw new Error(`Request returned ${x.status}`);
            setNodeInfo(x.body as NodeInfo);
        }).catch(e => {
            console.error("failed to GET /api/v1/node", e);
            setNodeInfo(null);
        });
    }, []);

    // Handle if the token is present.
    if (hasToken) {
        // Redirect to where is expected.
        const getLoginPath = () => {
            // Check for a redirect_to path in the query string.
            const params = new URLSearchParams(location.search);
            const redirectTo = params.get("redirect_to");
            if (redirectTo) {
                // Use this for the redirect.
                return redirectTo;
            }

            // Default to /app.
            return "/app";
        };
        return <Navigate replace to={getLoginPath()} />;
    }

    // Defines the main login layout.
    const setToken = (token: string) => {
        localStorage.setItem("token", token);
        setHasToken(true);
    };
    return <Login setToken={setToken} nodeInfo={nodeInfo} />;
};
