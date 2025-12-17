import "../../App.css";
import { invoke } from "@tauri-apps/api/core";
import { useEffect } from "react";
import { useNavigate } from "react-router-dom";

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
                sessionStorage.setItem("access_token", tokenData.access_token);
                sessionStorage.setItem("refresh_token", tokenData.refresh_token);

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