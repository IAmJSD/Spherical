<template>
    <div class="container has-text-centered">
        <h1 class="title is-1">Welcome to Spherical</h1>
        <h6 class="title is-6">The privacy focused, open source social network.</h6>
        <hr>
        <div class="notification is-danger" v-if="error !== ''">
            {{ error }}
        </div>
        <div v-if="register">
            <p>
                <a @click="() => {
                    this.$data.register = false;
                    this.$data.error = '';
                    this.$data.email = '';
                    this.$data.password = '';
                }">Do you want to sign in?</a>
            </p>
            <br>
            <p>
                <a class="button" @click="userRegister">
                    Register
                </a>
            </p>
        </div>
        <div v-else>
            <input class="input is-rounded" type="text" placeholder="E-mail Address" v-model="email">
            <br><br>
            <input class="input is-rounded" type="password" placeholder="Password" v-model="password">
            <br><br>
            <p>
                <a @click="() => {this.$data.register = true; this.$data.error = '';}">Do you want to sign up?</a>
            </p>
            <br>
            <p>
                <a class="button" @click="userLogin">
                    Login
                </a>
            </p>
        </div>
    </div>
</template>

<script>
    export default {
        name: "Login",
        data() {
            return {
                register: false,
                email: "",
                password: "",
                error: "",
            };
        },
        methods: {
            userRegister() {

            },
            userLogin() {
                const vm = this;
                this.login(this.$data.email, this.$data.password).then(() => {
                    if (vm.$data.signedInUser) {
                        // We are logged in.
                        vm.$data.email = "";
                        vm.$data.password = "";
                        vm.$data.error = "";
                    } else {
                        // We are *not* logged in.
                        vm.$data.error = "The e-mail address and password is not correct.";
                    }
                });
            },
        },
    }
</script>
