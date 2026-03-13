// Learn more about Tauri commands at https://tauri.app/develop/calling-rust/
use futures_util::{lock::Mutex, SinkExt, StreamExt};
use jsonwebtoken::Validation;
use jsonwebtoken::{decode, decode_header, TokenData};
use jsonwebtoken::{Algorithm, DecodingKey};
use keyring::Entry;
use serde::Deserialize;
use serde::Serialize;
use specta::Type;
use specta_typescript::Typescript;
use std::env;
use std::sync::{Arc, RwLock};
use tauri::{Emitter, State};
use tauri_specta::{collect_commands, Builder};
use tokio::sync::mpsc;
use tokio_tungstenite::connect_async;

#[derive(Debug, Serialize, Deserialize)]
struct JwkClaims {
    pub sub: String,
    pub exp: usize,
    pub iat: usize,
    pub iss: String,
    pub azp: String,
}

#[derive(Debug, Clone)]
struct JwksKey {
    kid: String,
    alg: Algorithm,
    decoding_key: DecodingKey, //this will be created from n and e
}

#[derive(Clone)]
struct AppState {
    jwks: Arc<RwLock<Vec<JwksKey>>>,
}

const BACKEND_REGISTER_URL: &str = env!("BACKEND_REGISTER_URL");
const REDIRECT_URL: &str = env!("REDIRECT_URL");
const TOKEN_URL: &str = env!("TOKEN_URL");
const CLIENT_ID: &str = env!("CLIENT_ID");
const BACKEND_URL: &str = std::env!("BACKEND_URL");

#[tauri::command]
#[specta::specta]
fn greet(name: &str) -> String {
    format!("Hello, {}! You've been greeted from Rust!", name)
}

struct WsSender(Mutex<Option<mpsc::UnboundedSender<String>>>);

// interface UserInfo {
//   avatar_url?: string;
//   name: string;
//   id: string;
//   email: string;
//   status: string;
// }

// export default UserInfo;

#[derive(Serialize, Deserialize, Type)]
pub struct UserInfo {
    pub id: String,
    pub avatar_url: Option<String>,
    pub email: String,
    pub username: String,
    pub status: String,
    pub created_at: String,
    pub updated_at: String,
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

#[derive(Serialize, Deserialize, Type)]
pub struct MessageResponse {
    pub id: String,
    pub owner_name: String,
    pub room_id: String,
    pub content: String,
    pub timeStamp: String,
}

#[derive(Serialize, Deserialize, Type)]
pub struct RoomLitePayload {
    pub id: String,
    pub name: String,
    pub description: String,
    pub created_at: String,
    pub updated_at: String,
}

//Fetch jwk code for validation
async fn fetch_public_code() -> Result<Vec<JwksKey>, String> {
    let url = "http://localhost:8081/realms/chat-app/protocol/openid-connect/certs";

    let client = reqwest::Client::new();
    let res = client.get(url).send().await.map_err(|e| e.to_string())?;

    if res.status().is_success() {
        //turn into json value
        let result = res.text().await.map_err(|e| e.to_string())?;
        let json_value: serde_json::Value =
            serde_json::from_str(&result).map_err(|e| e.to_string())?;

        let mut decoded_key = Vec::<JwksKey>::new();

        if let Some(key_array) = json_value["keys"].as_array() {
            for keys in key_array {
                if keys["use"] == "sig" {
                    let n = keys["n"].as_str().ok_or("Missing n")?;
                    let e = keys["e"].as_str().ok_or("Missing e")?;
                    let kid = keys["kid"].as_str().ok_or("Missing kid")?;

                    let decoding_key =
                        DecodingKey::from_rsa_components(n, e).map_err(|e| e.to_string())?;

                    decoded_key.push(JwksKey {
                        kid: kid.to_string(),
                        alg: Algorithm::RS256,
                        decoding_key,
                    });
                }
            }
        }
        println!("Decoded Key: {:?}", decoded_key);
        Ok(decoded_key)
    } else {
        let error_body = res
            .text()
            .await
            .unwrap_or_else(|_| "Cannot read error body".to_string());
        println!("{}", error_body);
        Err(format!("Error when fetching jwk: {}", error_body))
    }
}

async fn validate_token(
    token: &String,
    keys: Vec<JwksKey>,
) -> Result<TokenData<JwkClaims>, String> {
    let header = decode_header(token).map_err(|e| e.to_string())?;
    let kid = header.kid.ok_or("Token header is missing kid")?;

    let decoding_key = keys
        .iter()
        .find(|key| key.kid == kid)
        .ok_or("No matching public key was found")?;

    let mut validation = Validation::new(header.alg);
    validation.leeway = 60;
    validation.set_issuer(&["http://localhost:8081/realms/chat-app"]); // to make sure the token is from this specific endpoint
    validation.set_audience(&["account"]); //Get the client id from access token

    decode::<JwkClaims>(token, &decoding_key.decoding_key, &validation)
        .map_err(|e| format!("JWT validation failed: {}", e))
}

#[tauri::command]
#[specta::specta]
async fn fetch_token(code: String, verifier: String) -> Result<String, String> {
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
        let error_body = res
            .text()
            .await
            .unwrap_or_else(|_| "Cannot read error body".to_string());
        println!("{}", error_body);
        Err(format!("Error when fetching token: {}", error_body))
    }
}

#[tauri::command]
#[specta::specta]
async fn fetch_account_info(access_token: String) -> Result<UserInfo, String> {
    let client = reqwest::Client::new();

    let res = client
        .get("http://localhost:8081/realms/chat-app/protocol/openid-connect/userinfo")
        .bearer_auth(access_token)
        .send()
        .await
        .map_err(|e| e.to_string())?;

    if res.status().is_success() {
        let result = res.text().await.map_err(|e| e.to_string())?;
        //get the user email from the keycloak
        let user_data_keycloak: serde_json::Value =
            serde_json::from_str(&result).map_err(|e| e.to_string())?;

        let email = user_data_keycloak["email"]
            .as_str()
            .ok_or("Email not found in Keycloak response")?;

        //fetch the info from the db
        let url = format!("http://localhost:8080/fetch_user_info?username={}", email);
        let res = client.get(&url).send().await.map_err(|e| e.to_string())?;

        let user_info_db = res.text().await.map_err(|e| e.to_string())?;

        let user_info: UserInfo = serde_json::from_str(&user_info_db).map_err(|e| e.to_string())?;

        return Ok(user_info);
    } else {
        let error_body = res
            .text()
            .await
            .unwrap_or_else(|_| "Cannot read error body".to_string());
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
async fn finalize_login(access_token: String, refresh_token: String) -> bool {
    if save_data_to_keyring("access_token".to_string(), access_token.to_string()).is_ok()
        && save_data_to_keyring("refresh_token".to_string(), refresh_token.to_string()).is_ok()
    {
        return true;
    }

    if let Ok(access_token) = get_data_from_keyring("access_token".to_string()) {
        println!("Access Token from keyring: {}", access_token);
    }
    if let Ok(refresh_token) = get_data_from_keyring("refresh_token".to_string()) {
        println!("Refresh Token from keyring: {}", refresh_token);
    }
    return false;
}

#[tauri::command]
#[specta::specta]
async fn logout() -> bool {
    if (delete_data_in_keyring("access_token".to_string()).is_ok()
        && delete_data_in_keyring("refresh_token".to_string()).is_ok())
    {
        return true;
    }
    return false;
}

#[tauri::command]
#[specta::specta]
async fn checkAuth(state: tauri::State<'_, AppState>) -> Result<bool, String> {
    if let Ok(access_token) = get_data_from_keyring("access_token".to_string()) {
        let keys = state
            .jwks
            .read()
            .map_err(|_| "Something is wrong when fetching keys")?
            .clone();

        match validate_token(&access_token, keys).await {
            Ok(_) => {
                println!("Token is valid");
                Ok(true)
            }
            Err(e) => {
                println!("Token is invalid: {}", e);
                Ok(false)
            }
        }
    } else {
        return Ok(false);
    }
}

#[tauri::command]
#[specta::specta]
async fn establish_ws(
    app: tauri::AppHandle,
    state: State<'_, WsSender>,
    user_id: String,
) -> Result<(), String> {
    //NOTE: use wss for production which is more secured
    let url = format!("ws://localhost:8080/ws?user_id={}", user_id);

    let (ws_stream, _) = connect_async(&url).await.map_err(|e| e.to_string())?;

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
            if let Err(e) = write
                .send(tokio_tungstenite::tungstenite::Message::Text(msg.into()))
                .await
            {
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

#[tauri::command]
#[specta::specta]
async fn fetch_rooms_list(user_id: String) -> Result<Vec<RoomLitePayload>, String> {
    let client = reqwest::Client::new();
    let url = format!("{}/fetch_rooms?user_id={}", BACKEND_URL, user_id);
    let res = client.get(&url).send().await.map_err(|e| e.to_string())?;

    if res.status().is_success() {
        let text = res.text().await.map_err(|e| e.to_string())?;
        let rooms: Vec<RoomLitePayload> = serde_json::from_str(&text).map_err(|e| e.to_string())?;
        Ok(rooms)
        // Ok(res.text().await.map_err(|e| e.to_string())?)
    } else {
        let error_body = res
            .text()
            .await
            .unwrap_or_else(|_| "Cannot read error body".to_string());
        println!("{}", error_body);
        Err(format!("Error when fetching token: {}", error_body))
    }
}

#[tauri::command]
#[specta::specta]
async fn fetch_room_messages(room_id: String) -> Result<Vec<MessageResponse>, String> {
    let client = reqwest::Client::new();
    let url = format!("{}/fetch_room_message?room_id={}", BACKEND_URL, room_id);
    let res = client.get(&url).send().await.map_err(|e| e.to_string())?;

    if res.status().is_success() {
        let text = res.text().await.map_err(|e| e.to_string())?;
        println!("Message: {}", text);
        let rooms: Vec<MessageResponse> = serde_json::from_str(&text).map_err(|e| e.to_string())?;
        Ok(rooms)
    } else {
        let error_body = res
            .text()
            .await
            .unwrap_or_else(|_| "Cannot read error body".to_string());
        println!("{}", error_body);
        Err(format!("Error when fetching token: {}", error_body))
    }
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    let rt = tokio::runtime::Runtime::new().unwrap();
    let public_key = rt.block_on(fetch_public_code()).unwrap_or_else(|e| {
        println!("Failed to fetch token: {}", e);
        Vec::new()
    });

    let app_state = AppState {
        jwks: Arc::new(RwLock::new(public_key)),
    };

    dotenvy::dotenv().ok();

    let specta_builder = Builder::<tauri::Wry>::new().commands(collect_commands![
        greet,
        fetch_token,
        fetch_account_info,
        finalize_login,
        logout,
        get_data_from_keyring,
        checkAuth,
        establish_ws,
        send_message,
        fetch_rooms_list,
        fetch_room_messages
    ]);

    #[cfg(debug_assertions)] // <- Only export on non-release builds
    specta_builder
        .export(Typescript::default(), "../src/bindings.ts")
        .expect("Failed to export typescript bindings");

    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .plugin(tauri_plugin_store::Builder::default().build())
        .manage(WsSender(Mutex::new(None)))
        .manage(app_state)
        .invoke_handler(specta_builder.invoke_handler())
        .setup(move |app| {
            specta_builder.mount_events(app);
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
