// Learn more about Tauri commands at https://tauri.app/develop/calling-rust/
use dotenv::dotenv;
use std::env;

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

    let register_url = env::var("BACKEND_REGISTER_URL").expect("BACKEND_REGISTER_URL must be set in .env");

    let client = reqwest::Client::new();

    let res = client
            .post(&register_url)
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
    let token_url = env::var("TOKEN_URL").expect("TOKEN_URL must be set in .env");
    let redirect_url = env::var("REDIRECT_URL").expect("REDIRECT_URL must be set in .env");
    let client_id = env::var("CLIENT_ID").expect("CLIENT_ID must be set in .env");

    let params = [
        ("grant_type", "authorization_code"),
        ("client_id", &client_id),
        ("code", &code),
        ("redirect_uri", &redirect_url),
        ("code_verifier", &verifier),
    ];

    let res = client
            .post(&token_url)
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

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    dotenv().ok();
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .invoke_handler(tauri::generate_handler![greet, register_account, fetch_token])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
