// Imports needed stuff.
import axios from "axios"
import Cryptr from "cryptr"
// @ts-ignore
import * as toUint8ArrayDef from "base64-to-uint8array"
const toUint8Array = toUint8ArrayDef as (data: string) => Uint8Array

// The user API interface.
export class APIUser {
    // Defines the decryption key, private key and token.
    public cryptoHandler: Cryptr
    public privateKey: Uint8Array
    public token: string
    public profile: Promise<{
        firstName: string;
        lastName: string;
        email: string;
        emailConfirmed: boolean;
        description: string;
        phoneHash: string | undefined;
        twoFactor: boolean;
        profilePicture: string;
        publicKey: string;
        createdAt: number;
    }> = Promise.reject("User profile is unset. Initialisation went wrong!")

    // Constructs the class.
    public constructor(cryptoHandler: Cryptr, privateKey: Uint8Array, token: string) {
        this.cryptoHandler = cryptoHandler
        this.privateKey = privateKey
        this.token = token
    }

    // Logs out of Spherical.
    public static async logout() {
        // Logs out.
        await axios.get("/api/v1/user/logout", {
            headers: {
                Token: localStorage.getItem("token"),
            },
        })

        // Removes the token and decryption key.
        localStorage.removeItem("decryption_key")
        localStorage.removeItem("token")

        // Set the signed in user to null.
        signedInUser = null
    }

    // Logs into Spherical.
    public static async login(email: string, password: string): Promise<APIUser | null> {
        // TODO: handle 2FA.

        // Gets the decryption key for the private key.
        const decryptionKey = (await crypto.subtle.digest("SHA-256", new TextEncoder(
            // Idk why TS is pissed off here.
            // @ts-ignore
            "utf-8"
        ).encode(password)).then(buf => {
            return Array.prototype.map.call(new Uint8Array(buf), (x: any) => ((`00${x.toString(16)}`).slice(-2)))
        })).substr(0, 32)

        // Creates a SHA-512 string of the password.
        const sha512Password = (await crypto.subtle.digest("SHA-512", new TextEncoder(
            // Idk why TS is pissed off here.
            // @ts-ignore
            "utf-8"
        ).encode(password)).then(buf => {
            return Array.prototype.map.call(new Uint8Array(buf), (x: any) => ((`00${x.toString(16)}`).slice(-2)))
        }))

        // Gets the token from the API if all the credentials are correct.
        let token: string
        try {
            const response = await axios.get("/api/v1/user/auth", {
                headers: {
                    Email: email,
                    Password: sha512Password,
                },
                validateStatus: status => {
                    return status >= 200 && status < 300
                },
                responseType: "json",
            })
            token = response.data as string
        } catch (_) {
            return null
        }

        // Writes the decryption key and token to local storage.
        localStorage.setItem("decryption_key", decryptionKey)
        localStorage.setItem("token", token)

        // Does the remainder of the login.
        const user = this.browserLogin()
        signedInUser = user
        return user
    }

    // Logs in from the browser information.
    public static async browserLogin(): Promise<APIUser | null> {
        // Gets the token and decryption key.
        const decryptionKey = localStorage.getItem("decryption_key")
        const token = localStorage.getItem("token")

        // If one of them is null, return here.
        if (!token || !decryptionKey) return

        // Gets the encrypted private key and decrypt it.
        const encryptedKeyResponse = await axios.get("/api/v1/user/private_key", {
            headers: {
                Token: token,
            },
            validateStatus: status => {
                return status >= 200 && status < 300
            },
            responseType: "json",
        })
        const cryptr = new Cryptr(decryptionKey)
        const privateKey = toUint8Array(cryptr.decrypt(encryptedKeyResponse.data))

        // Creates a user class.
        const user = new APIUser(cryptr, privateKey, token)
        user.profile = await axios.get("/api/v1/user/profile", {
            headers: {
                Token: token,
            },
            validateStatus: status => {
                return status >= 200 && status < 300
            },
            responseType: "arraybuffer",
        }).then(r => r.data)
        return user
    }
}

// The currently signed in user.
let signedInUser: Promise<null | APIUser> = APIUser.browserLogin()

// Returns the user.
export default signedInUser
