<template>
    <div v-if="!$props.hidden">
        <h1>navbar</h1>
    </div>
</template>

<script>
    export default {
        name: "Navbar",
        props: {
            hidden: Boolean,
        },
    }
</script>
