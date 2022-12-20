import { h } from "preact";
import { useEffect, useState } from "preact/hooks";
import { Navigate, useLocation } from "react-router-dom";
import LoadingSplash from "./LoadingSplash";
import {
    addDisconnectHandler, addReadyHandler,
    removeDisconnectHandler, removeReadyHandler,
    startWebsocket,
} from "../../helpers/gateway";

export default () => {
    // Get the location information.
    const location = useLocation();

    // Handle displaying a loading screen whilst everything is loading.
    const [state, setState] = useState(1);
    useEffect(() => {
        // If the user isn't logged in, redirect them to the login page.
        if (!localStorage.getItem("token")) {
            setState(0);
            return;
        }

        // Add a disconnect handler to the websocket.
        const disconnectHdId = addDisconnectHandler(reconnect => {
            setState(reconnect ? 1 : 0);
        });

        // Remove the ready handler.
        const readyHnId = addReadyHandler(() => {
            setState(2);
        })

        // Start the websocket up.
        startWebsocket();

        // Return a function to remove the handlers.
        return () => {
            removeDisconnectHandler(disconnectHdId);
            removeReadyHandler(readyHnId);
        };
    }, []);

    // Handle state 0 which is go to the login page.
    if (state === 0) return <Navigate replace to={
        `/?redirect_to=${encodeURIComponent(location.pathname)}`} />;

    // Handle state 1 which is loading.
    if (state === 1) return <LoadingSplash />;

    // Render the app base.
    return <div>
        <h1>Connected!</h1>
        <p>Check console.</p>
    </div>;
};
