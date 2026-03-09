// Learn more about Tauri commands at https://tauri.app/develop/calling-rust/
use std::{env, fs::remove_file, string};
use serde_json::json;
use tauri_plugin_store::StoreExt;
use keyring::{Entry};
use futures_util::{SinkExt, StreamExt, lock::Mutex};
use tauri::{Emitter, Manager, State};
use serde::Serialize;
use tokio::sync::mpsc;
use tokio_tungstenite::connect_async;
use tauri_specta::{collect_commands, Builder};
use serde::Deserialize;
use specta::Type;
use specta_typescript::Typescript;

const BACKEND_REGISTER_URL: &str = env!("BACKEND_REGISTER_URL");
const REDIRECT_URL: &str = env!("REDIRECT_URL");
const TOKEN_URL: &str = env!("TOKEN_URL");
const CLIENT_ID: &str = env!("CLIENT_ID");

#[tauri::command]
#[specta::specta]
fn greet(name: &str) -> String {
    format!("Hello, {}! You've been greeted from Rust!", name)
}


struct WsSender(Mutex<Option<mpsc::UnboundedSender<String>>>);

#[derive(Serialize)]
struct UserAccountPayload {
    #[serde(rename="FirstName")]
    first_name: String,
    #[serde(rename="LastName")]
    last_name: String,
    password: String,
    email: String,
}

#[derive(Serialize, Deserialize, Type)]
pub struct MessagePayload {
    pub id: String,
    pub user_id: String,
    pub room_id: String,
    pub content: String,
    pub timeStamp: String,
    pub action: String,
}

#[tauri::command]
#[specta::specta]
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
#[specta::specta]
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
#[specta::specta]
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

fn save_data_to_keyring(name: String, data: String) -> Result<(), keyring::Error> {
    let entry = Entry::new("chat-app-service", &name)?;
    entry.set_password(&data)
}

#[tauri::command]
#[specta::specta]
fn get_data_from_keyring(name: String) -> Result<String, String> {
    let entry = Entry::new("chat-app-service", &name).map_err(|e| e.to_string())?;
    entry.get_password().map_err(|e| e.to_string())
}

fn delete_data_in_keyring(name: String) -> Result<(), keyring::Error> {
    let entry = Entry::new("chat-app-service", &name)?;
    entry.delete_credential()?;
    Ok(())
}
 
#[tauri::command]
#[specta::specta]
async fn login(access_token: String, refresh_token: String) -> bool {
    if save_data_to_keyring("access_token".to_string(), access_token.to_string()).is_ok() &&
        save_data_to_keyring("refresh_token".to_string(), refresh_token.to_string()).is_ok() {
        return true;
    }

    if let Ok(access_token) = get_data_from_keyring("access_token".to_string()) {
        println!("Access Token from keyring: {}", access_token);
    }
    if let Ok(refresh_token) = get_data_from_keyring("refresh_token".to_string()) {
        println!("Refresh Token from keyring: {}", refresh_token);
    }
    return false
} 

#[tauri::command]
#[specta::specta]
async fn logout() -> bool {
    if(delete_data_in_keyring("access_token".to_string()).is_ok() && 
        delete_data_in_keyring("refresh_token".to_string()).is_ok()) {
        return true;
    }
    return false;
}

#[tauri::command]
#[specta::specta]
async fn checkAuth() -> bool {
    if let Ok(access_token) = get_data_from_keyring("access_token".to_string()) {
        let is_ok = fetch_account_info(access_token).await.is_ok();
        println!("test: {}", is_ok);
        return is_ok
    } else {
        return false
    }
}

#[tauri::command]
#[specta::specta]
async fn establish_ws(app: tauri::AppHandle, state: State<'_, WsSender>, user_id: String) -> Result<(), String> {
    //NOTE: use wss for production which is more secured
    let url = format!("ws://localhost:8080/ws?user_id={}", user_id);

    let (ws_stream, _) = connect_async(&url)
                            .await
                            .map_err(|e| e.to_string())?;
    

    let (mut write, mut read) = ws_stream.split();
    
    let (tx, mut rx) = mpsc::unbounded_channel::<String>();

    {
        let mut sender = state.0.lock().await;
        *sender = Some(tx);
    }

    let app_handle = app.clone();

    //Reading message from server via WS
    //async move make the ws be able to continue to run even after function return
    tokio::spawn(async move {
        while let Some(msg) = read.next().await {
            if let Ok(msg) = msg {
                let _ = app_handle.emit("ws-message", msg.to_text().unwrap_or(""));
            }
        }
    });

    //Sending message to server
    tokio::spawn(async move {
        while let Some(msg) = rx.recv().await {
            if let Err(e) = write.send(tokio_tungstenite::tungstenite::Message::Text(msg.into())).await {
                println!("Error when sending message: {}", e);
                break;
            }      
        }
    });

    Ok(())
}

#[tauri::command]
#[specta::specta]
async fn send_message(state: State<'_, WsSender>, message: MessagePayload) -> Result<(), String> {
    //lock the mutex
    let sender = state.0.lock().await;
    if let Some(tx) = sender.as_ref() {
        let json = serde_json::to_string(&message).map_err(|e| e.to_string())?;
        tx.send(json).map_err(|e| e.to_string())?;
        Ok(())
    } else {
        Err("Something is wrong with websocker connection".to_string())
    }
}

// #[tauri::command]
// async fn request_message(room_id: string) -> Result<(), String> {
    
// }

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    dotenvy::dotenv().ok();

    let mut specta_builder = Builder::<tauri::Wry>::new()
        .commands(collect_commands![
            greet,
            register_account,
            fetch_token,
            fetch_account_info,
            login,
            logout,
            get_data_from_keyring,
            checkAuth,
            establish_ws,
            send_message,
        ]);

    #[cfg(debug_assertions)] // <- Only export on non-release builds
    specta_builder
        .export(Typescript::default(), "../src/bindings.ts")
        .expect("Failed to export typescript bindings");

    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .plugin(tauri_plugin_store::Builder::default().build())
        .manage(WsSender(Mutex::new(None)))
        .invoke_handler(specta_builder.invoke_handler()) 
        .setup(move |app| {
            specta_builder.mount_events(app);
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
