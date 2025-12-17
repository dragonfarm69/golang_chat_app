import { invoke } from "@tauri-apps/api/core";
import "../../App.css";
import { useState } from "react";

function generateCodeVerifier() {
    const array = new Uint8Array(32);
    window.crypto.getRandomValues(array);
    return btoa(String.fromCharCode(...array))
        .replace(/\+/g, '-')
        .replace(/\//g, '_')
        .replace(/=+$/, '');
}

async function generateCodeChallenge(verifier: string) {
    const encoder = new TextEncoder();
    const data = encoder.encode(verifier);

    const digest = await crypto.subtle.digest("SHA-256", data);
    return btoa(String.fromCharCode(...new Uint8Array(digest)))
        .replace(/\+/g, '-')
        .replace(/\//g, '_')
        .replace(/=+$/, '');
}

function LoginPage() {

    const handleOnClick = async () => {
        const verifier = generateCodeVerifier();
        const challenge = await generateCodeChallenge(verifier);

        //Verifier to exchange token
        sessionStorage.setItem("pkce_verifier", verifier);

        const params = new URLSearchParams({
            client_id: "authentication-cli",
            response_type: "code",
            scope: "openid profile email",
            redirect_uri: "http://localhost:1420/callback",
            code_challenge: challenge,
            code_challenge_method: "S256",
        });


        window.location.href =
            "http://localhost:8081/realms/chat-app/protocol/openid-connect/auth?" +
            params.toString();
    }

    return (
        <div className="app-container" style={{backgroundColor: "white"}}>
            <button className="join-room-button" type="button" onClick={handleOnClick}>Login</button>
        </div>
    )
}

export default LoginPage