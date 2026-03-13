import "../../App.css";
import { invoke } from "@tauri-apps/api/core";
import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Store } from "@tauri-apps/plugin-store";
import { useAuth } from "../../Context/AuthContext";
import { useUser } from "../../Context/userContext";
import { commands } from "../../bindings";
import { UserInfo } from "../../bindings";

function CallBackPage() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const { initializeUserContext, userData } = useUser();

  useEffect(() => {
    const finalizeLogin = async () => {
      try {
        const verifier = sessionStorage.getItem("pkce_verifier");

        const current_url_params = new URLSearchParams(window.location.search);
        const code = current_url_params.get("code");

        const response = await invoke("fetch_token", {
          verifier: verifier,
          code: code,
        });

        const tokenData = JSON.parse(response as string);

        sessionStorage.removeItem("pkce_verifier");

        //store tokens
        await login(tokenData.access_token, tokenData.refresh_token);

        const is_valid_token = await commands.checkAuth();

        if (!is_valid_token) {
          navigate("/error");
          return;
        }

        const userInfo = await commands.fetchAccountInfo(
          tokenData.access_token,
        );

        let extractedUserData;
        if (userInfo.status === "ok") {
          extractedUserData = userInfo.data;
        }

        // //store info in store
        const store = await Store.load("cache_data.json");
        await store.set("email", extractedUserData?.email);
        await store.set("username", extractedUserData?.username);
        await store.save(); // persist the store to disk

        await initializeUserContext(extractedUserData as UserInfo);

        //this should lead to home page
        navigate("/");
      } catch (error) {
        console.log(error);
        alert(`Login failed ${error}`);
      }
    };

    finalizeLogin();
  }, [navigate]);

  return (
    <div className="app-container" style={{ backgroundColor: "white" }}>
      <h1>Finalize login...</h1>
    </div>
  );
}

export default CallBackPage;
