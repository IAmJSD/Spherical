import "preact/devtools";

// Remove the setup_stickies local storage item.
localStorage.removeItem("setup_stickies");

import { h, render } from "preact";
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import Root from "./components/main/routes/Root";
import AppBootstrap from "./components/main/AppBootstrap";

const router = createBrowserRouter([
    {
        path: "/",
        element: <Root />,
    },
    {
        path: "/app",
        element: <AppBootstrap />,
        children: [],
    },
]);

render(<RouterProvider router={router} />, document.getElementById("app_mount"));
