<template>
    <BaseComponent :navbar-hidden="!signedInUser">
        <span v-if="signedInUser">

        </span>
        <span>
            <Login></Login>
        </span>
    </BaseComponent>
</template>

<script>
    import BaseComponent from "../components/Base.vue"
    import Login from "../components/Login.vue"

    export default {
        name: "Home",
        components: {
            BaseComponent,
            Login,
        },
    }
</script>
