<template>
    <div v-if="!hidden">
        <nav class="navbar is-dark">
            <div class="navbar-start">
                <div class="navbar-brand">
                    <router-link :to="{name: 'Home', params: {}}" class="navbar-item">
                        <b>Spherical</b>
                    </router-link>
                </div>
            </div>
            <div class="navbar-end">
                <div class="dropdown" v-if="signedInUser">
                    <div class="dropdown-trigger">
                        <button class="button" aria-haspopup="true" aria-controls="dropdown-menu3">
                            <span>{{ profile.firstName }}</span>
                            <span class="icon is-small">
                                <i class="fas fa-angle-down" aria-hidden="true"></i>
                            </span>
                        </button>
                    </div>
                    <div class="dropdown-menu" role="menu">
                        <div class="dropdown-content">
                            <router-link :to="{name: 'Profile', params: {'uid': profile.id}}" class="navbar-item">
                                Profile
                            </router-link>
                        </div>
                    </div>
                </div>
                <div class="navbar-brand" v-else>
                    <router-link :to="{name: 'Home', params: {}}" class="navbar-item">
                        Login
                    </router-link>
                </div>
            </div>
        </nav>
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
