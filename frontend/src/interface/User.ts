// Imports needed stuff.
import axios from "axios"
import Vue from "vue"

// The user API interface.
export class APIUser {
    // Defines the private key and token.
    private token: string;
    public profile: Promise<{
        firstName: string;
        lastName: string;
        email: string;
        emailConfirmed: boolean;
        description: string;
        phoneHash: string | undefined;
        twoFactor: boolean;
        profilePicture: string;
        createdAt: number;
        id: string;
    }> = Promise.reject("User profile is unset. Initialisation went wrong!");

    // Constructs the class.
    public constructor(token: string) {
        this.token = token;
    }

    // Logs out of Spherical.
    public static async logout() {
        // Logs out.
        await axios.get("/api/v1/user/logout", {
            headers: {
                Token: localStorage.getItem("token"),
            },
        });

        // Removes the token.
        localStorage.removeItem("token");
    }

    // Logs into Spherical.
    public static async login(email: string, password: string): Promise<APIUser | null> {
        // TODO: handle 2FA.

        // Creates a SHA-512 string of the password.
        const sha512Password = (await crypto.subtle.digest("SHA-512", new TextEncoder(
            // Idk why TS is pissed off here.
            // @ts-ignore
            "utf-8"
        ).encode(password)).then(buf => {
            return Array.prototype.map.call(new Uint8Array(buf), (x: any) => ((`00${x.toString(16)}`).slice(-2)));
        }));

        // Gets the token from the API if all the credentials are correct.
        let token: string;
        try {
            const response = await axios.get("/api/v1/user/auth", {
                headers: {
                    Email: email,
                    Password: sha512Password,
                },
                validateStatus(status: number) {
                    return status >= 200 && status < 300
                },
                responseType: "json",
            });
            token = response.data as string;
        } catch (_) {
            return null;
        }

        // Writes the token to local storage.
        localStorage.setItem("token", token);

        // Does the remainder of the login.
        return this.browserLogin();
    }

    // Logs in from the browser information.
    public static async browserLogin(): Promise<APIUser | null> {
        // Gets the token.
        const token = localStorage.getItem("token");

        // If the token is null, return here.
        if (!token) return;

        // Creates a user class.
        const user = new APIUser(token);
        user.profile = await axios.get("/api/v1/user/profile", {
            headers: {
                Token: token,
            },
            validateStatus(status: number) {
                return status >= 200 && status < 300
            },
            responseType: "json",
        }).then((r: any) => r.data);
        return user;
    }
}

// Set the signed in user as a mixin.
Vue.mixin({
    data() {
        return {
            signedInUser: null,
            profile: {},
        };
    },
    methods: {
        login(email: string, password: string) {
            const vm = this;
            return APIUser.login(email, password).then(result => {
                vm.$data.signedInUser = result;
                if (result) {
                    result.profile.then((p: any) => {
                        vm.$data.profile = p;
                    });
                }
            });
        },
        logout() {
            const vm = this;
            return this.$data.signedInUser.logout().then(() => {
                vm.$data.signedInUser = null;
                vm.$data.profile = {};
            });
        },
    },
    created() {
        const vm = this;
        APIUser.browserLogin().then(result => {
            vm.$data.signedInUser = result;
            if (result) {
                result.profile.then((p: any) => {
                    vm.$data.profile = p;
                });
            }
        });
    },
});
