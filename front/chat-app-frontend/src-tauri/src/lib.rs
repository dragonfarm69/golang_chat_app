// Learn more about Tauri commands at https://tauri.app/develop/calling-rust/
use std::env;

const BACKEND_REGISTER_URL: &str = env!("BACKEND_REGISTER_URL");
const REDIRECT_URL: &str = env!("REDIRECT_URL");
const TOKEN_URL: &str = env!("TOKEN_URL");
const CLIENT_ID: &str = env!("CLIENT_ID");

#[tauri::command]
fn greet(name: &str) -> String {
    format!("Hello, {}! You've been greeted from Rust!", name)
}

use serde::Serialize;

#[derive(Serialize)]
struct UserAccountPayload {
    #[serde(rename="FirstName")]
    first_name: String,
    #[serde(rename="LastName")]
    last_name: String,
    password: String,
    email: String,
}

#[tauri::command]
async fn register_account(first_name: String, last_name: String, password: String, email: String) -> Result<String, String> {
    //create user payload
    let payload = UserAccountPayload{
        first_name,
        last_name,
        password,
        email
    };

    let client = reqwest::Client::new();

    let res = client
            .post(BACKEND_REGISTER_URL)
            .json(&payload)
            .send()
            .await
            .map_err(|e| e.to_string())?;
    
    if res.status().is_success() {
        Ok(res.text().await.map_err(|e| e.to_string())?)
    } else {
        Err(format!("Server responsded with error: {}", res.status()))
    }
}

#[tauri::command]
async fn fetch_token(code: String, verifier: String) -> Result<String, String>{
    let client = reqwest::Client::new();

    let params = [
        ("grant_type", "authorization_code"),
        ("client_id", CLIENT_ID),
        ("code", &code),
        ("redirect_uri", REDIRECT_URL),
        ("code_verifier", &verifier),
    ];

    let res = client
            .post(TOKEN_URL)
            .form(&params) //send data as x-www-form-urlencoded
            .send()
            .await
            .map_err(|e| e.to_string())?;

    if res.status().is_success() {
        Ok(res.text().await.map_err(|e| e.to_string())?)
    } else {
        let error_body = res.text().await.unwrap_or_else(|_| "Cannot read error body".to_string());
        println!("{}", error_body);
        Err(format!("Error when fetching token: {}", error_body))
    }
}

#[tauri::command]
async fn fetch_account_info(access_token: String) -> Result<String, String> {
    let client = reqwest::Client::new();

    let res = client
            .get("http://localhost:8081/realms/chat-app/protocol/openid-connect/userinfo")
            .bearer_auth(access_token)
            .send()
            .await
            .map_err(|e| e.to_string())?;

    if res.status().is_success() {
        Ok(res.text().await.map_err(|e| e.to_string())?)
    } else {
        let error_body = res.text().await.unwrap_or_else(|_| "Cannot read error body".to_string());
        println!("{}", error_body);
        Err(format!("Error when fetching data: {}", error_body))
    }
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    dotenvy::dotenv().ok();
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .plugin(tauri_plugin_store::Builder::default().build())
        .invoke_handler(tauri::generate_handler![greet, register_account, fetch_token, fetch_account_info])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
