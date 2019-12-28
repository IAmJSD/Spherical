<template>
    <div class="spherical-base">
        <Navbar :hidden="navbarHidden"></Navbar>
        <slot></slot>
    </div>
</template>

<script>
    import Navbar from "./Navbar.vue"

    export default {
        name: "BaseComponent",
        components: {
            Navbar,
        },
        props: {
            navbarHidden: Boolean,
        },
    }
</script>
