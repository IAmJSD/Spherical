<template>
    <BaseComponent :navbar-hidden="true">
        <h1 class="title">Hello World!</h1>
    </BaseComponent>
</template>

<script>
    import BaseComponent from "../components/Base.vue"

    export default {
        name: "Home",
        components: {
            BaseComponent,
        },
    }
</script>
