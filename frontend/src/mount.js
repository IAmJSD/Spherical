// Imports everything needed for initialisation.
import Vue from "vue"
import VueRouter from "vue-router"

// Gets the payload.
import "./payload"

// Imports all of the used routes.
import Home from "./routes/Home.vue"
import App from "./components/App.vue"

// Tells Vue to use the router.
Vue.use(VueRouter);

// Creates the router.
const router = new VueRouter({
    mode: "history",
    routes: [
        {
            path: "/",
            name: "Home",
            component: Home,
        },
    ],
});

// Creates the Vue instance.
new Vue({
    el: "#app",
    router,
    render: x => x(App),
});
