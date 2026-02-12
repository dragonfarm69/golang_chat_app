import "../../App.css";
import { invoke } from "@tauri-apps/api/core";
import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Store } from "@tauri-apps/plugin-store";

function CallBackPage() {
    const navigate = useNavigate();
    useEffect(() => {
        const finalizeLogin = async () => {
            try {
                const verifier = sessionStorage.getItem("pkce_verifier")
    
                const current_url_params = new URLSearchParams(window.location.search)
                const code = current_url_params.get("code");
                const response = await invoke("fetch_token", {verifier, code})
    
                const tokenData = JSON.parse(response as string);
    
                sessionStorage.removeItem("pkce_verifier");
                
                //store tokens
                const loginSuccess = await invoke("login", { verifier, code });

                if (!loginSuccess) {
                    throw new Error("Failed to store tokens securely.");
                }


                const userInfo = await invoke("fetch_account_info", { accessToken: tokenData.access_token })
                const userData = JSON.parse(userInfo as string)

                //store info in store 
                const store = await Store.load("cache_data.json");
                await store.set("email", userData.email);
                await store.set("username", userData.name);
                await store.save(); // persist the store to disk
                navigate("/join");
            }
            catch (error) {
                console.log(error)
                alert(`Login failed ${error}`);
            }
        }

        finalizeLogin();
    }, [navigate])

    return (
        <div className="app-container" style={{backgroundColor: "white"}}>
            <h1>Finalize login...</h1>
        </div>
    )
}

export default CallBackPage