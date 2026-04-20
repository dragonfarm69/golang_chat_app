use chrono::{DateTime, Utc};
// Learn more about Tauri commands at https://tauri.app/develop/calling-rust/
use base64::{engine::general_purpose::URL_SAFE_NO_PAD, Engine as _};
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
use tokio::sync::{mpsc, oneshot};
use tokio::time::{interval, Duration};
use tokio_tungstenite::connect_async;

#[derive(Debug, Serialize, Deserialize)]
struct TokenResponse {
    refresh_token: String,
    access_token: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct TokenClaims {
    pub exp: usize,
    pub iss: String,
}

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
    pub stop_tx: Arc<Mutex<Option<oneshot::Sender<()>>>>,
}

const BACKEND_REGISTER_URL: &str = env!("BACKEND_REGISTER_URL");
const REDIRECT_URL: &str = env!("REDIRECT_URL");
const TOKEN_URL: &str = env!("TOKEN_URL");
const CLIENT_ID: &str = env!("CLIENT_ID");
const CLIENT_SECRET: &str = env!("CLIENT_SECRET");
const BACKEND_URL: &str = std::env!("BACKEND_URL");
const BACKEND_REFRESH_TOKEN_URL: &str = std::env!("BACKEND_REFRESH_TOKEN");

#[tauri::command]
#[specta::specta]
fn greet(name: &str) -> String {
    format!("Hello, {}! You've been greeted from Rust!", name)
}

struct WsSender(Mutex<Option<mpsc::UnboundedSender<String>>>);

#[derive(Serialize, Deserialize, Type)]
pub struct RegisterPayload {
    pub email: String,
    pub first_name: String,
    pub last_name: String,
    pub password: String,
}

#[derive(Serialize, Deserialize, Type, Debug)]
pub struct UserInfo {
    pub id: String,
    pub avatar_url: Option<String>,
    pub email: String,
    pub username: String,
    pub status: String,
    pub created_at: String,
    pub updated_at: Option<String>,
}

#[derive(Serialize, Deserialize, Type)]
pub struct MessagePayload {
    pub id: String,
    pub user_id: String,
    pub room_id: String,
    pub content: String,
    pub username: String,
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

#[derive(Serialize, Deserialize, Type, Debug)]
pub struct JoinRoomPayload {
    user_id: String,
    room_id: String,
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

    println!("Token for validating: {}", token);

    decode::<JwkClaims>(token, &decoding_key.decoding_key, &validation)
        .map_err(|e| format!("JWT validation failed: {}", e))
}

#[tauri::command]
#[specta::specta]
async fn refresh_token() -> Result<TokenResponse, String> {
    let client = reqwest::Client::new();

    let refresh_token = get_data_from_keyring("refresh_token".to_string())?;

    let res = client
        .post(BACKEND_REFRESH_TOKEN_URL)
        .json(&serde_json::json!({ "refresh_token": &refresh_token }))
        .send()
        .await
        .map_err(|e| e.to_string())?;

    if res.status().is_success() {
        let body = res.text().await.map_err(|e| e.to_string())?;
        let data = serde_json::from_str::<TokenResponse>(&body).map_err(|e| e.to_string())?;
        Ok(data)
    } else {
        let error_body = res
            .text()
            .await
            .unwrap_or_else(|_| "Cannot read error body".to_string());
        println!("{}", error_body);
        Err(format!(
            "Error when trying to refresh token: {}",
            error_body
        ))
    }
}

#[tauri::command]
#[specta::specta]
async fn fetch_token(code: String, verifier: String) -> Result<String, String> {
    let client = reqwest::Client::new();

    let params = [
        ("grant_type", "authorization_code"),
        ("client_id", CLIENT_ID),
        ("client_secret", CLIENT_SECRET),
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
        println!("Token fetched successfully");
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
async fn fetch_account_info() -> Result<UserInfo, String> {
    let access_token = get_data_from_keyring("access_token".to_string())?;
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

        let username = user_data_keycloak["preferred_username"]
            .as_str()
            .ok_or("username not found in Keycloak response")?;

        //fetch the info from the db
        let url = format!(
            "http://localhost:8080/api/fetch_user_info?username={}",
            username
        );
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
async fn finalize_login(
    state: tauri::State<'_, AppState>,
    access_token_string: String,
    refresh_token_string: String,
) -> Result<bool, String> {
    let stop_tx = state.stop_tx.clone();
    drop(state);

    if let Some(tx) = stop_tx.lock().await.take() {
        let _ = tx.send(());
    }

    if save_data_to_keyring("access_token".to_string(), access_token_string.to_string()).is_ok()
        && save_data_to_keyring(
            "refresh_token".to_string(),
            refresh_token_string.to_string(),
        )
        .is_ok()
    {
        let (tx, mut rx) = oneshot::channel::<()>();
        *stop_tx.lock().await = Some(tx);

        // create a timer for refreshing token
        tauri::async_runtime::spawn(async move {
            let mut ticker = interval(Duration::from_secs(300));
            ticker.tick().await; // skip the first timer

            loop {
                tokio::select! {
                    _ = &mut rx => break,
                    _ = ticker.tick() => {
                        match refresh_token().await {
                            Ok(new_token) => {
                                let _ = save_data_to_keyring("access_token".to_string(), new_token.access_token);
                                let _ = save_data_to_keyring("refresh_token".to_string(), new_token.refresh_token);
                            }
                            Err(e) => {
                                println!("Error when trying to refresh token {}", e);
                                break;
                            }
                        }
                    }
                }
            }
        });
        return Ok(true);
    }

    return Ok(false);
}

#[tauri::command]
#[specta::specta]
async fn register(data: RegisterPayload) -> Result<bool, String> {
    let client = reqwest::Client::new();

    let res = client
        .post(BACKEND_REGISTER_URL)
        .json(&serde_json::json!(data))
        .send()
        .await
        .map_err(|e| e.to_string())?;

    if (res.status().is_success()) {
        return Ok(true);
    }
    return Ok(false);
}

#[tauri::command]
#[specta::specta]
async fn logout() -> Result<bool, String> {
    let client = reqwest::Client::new();

    let res = client
        .post(BACKEND_URL.to_owned() + "/api/logout")
        .send()
        .await
        .map_err(|e| e.to_string())?;

    if (res.status().is_success()) {
        println!("User black listed")
    }

    if delete_data_in_keyring("access_token".to_string()).is_ok()
        && delete_data_in_keyring("refresh_token".to_string()).is_ok()
    {
        return Ok(true);
    }
    Ok(false)
}

#[tauri::command]
#[specta::specta]
async fn checkLoginStatus(state: tauri::State<'_, AppState>) -> Result<bool, String> {
    if let Ok(access_token) = get_data_from_keyring("access_token".to_string()) {
        let keys = state
            .jwks
            .read()
            .map_err(|_| "Something is wrong when fetching keys")?
            .clone();

        let accesss_token_clone = access_token.clone();

        let payload_b64 = accesss_token_clone.split('.').nth(1).ok_or("invalid jwt")?;

        let payload = URL_SAFE_NO_PAD
            .decode(payload_b64)
            .map_err(|e| e.to_string())?;

        let claims: TokenClaims = serde_json::from_slice(&payload).map_err(|e| e.to_string())?;
        let current_time: DateTime<Utc> = Utc::now();

        let exp_i64: i64 = claims
            .exp
            .try_into()
            .map_err(|e: std::num::TryFromIntError| e.to_string())?;

        let expire_time =
            DateTime::from_timestamp(exp_i64, 0).ok_or_else(|| "invalid timeStamp".to_string())?;

        if expire_time < current_time {
            return Ok(false);
        }

        match validate_token(&access_token, keys).await {
            Ok(_) => {
                println!("True");
                Ok(true)
            }
            Err(_) => {
                println!("Failed");
                Ok(false)
            }
        }
    } else {
        return Ok(false);
    }
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

        let accesss_token_clone = access_token.clone();

        let payload_b64 = accesss_token_clone.split('.').nth(1).ok_or("invalid jwt")?;
        let payload = URL_SAFE_NO_PAD
            .decode(payload_b64)
            .map_err(|e| e.to_string())?;
        let claims: TokenClaims = serde_json::from_slice(&payload).map_err(|e| e.to_string())?;
        let exp_i64: i64 = claims
            .exp
            .try_into()
            .map_err(|e: std::num::TryFromIntError| e.to_string())?;

        let expire_time =
            DateTime::from_timestamp(exp_i64, 0).ok_or_else(|| "invalid timeStamp".to_string())?;

        let current_time: DateTime<Utc> = Utc::now();

        let access_token = if expire_time < current_time {
            match refresh_token().await {
                Ok(result) => {
                    let _ = save_data_to_keyring(
                        "access_token".to_string(),
                        result.access_token.clone(),
                    );
                    let _ = save_data_to_keyring("refresh_token".to_string(), result.refresh_token);
                    result.access_token
                }
                Err(_) => return Ok(false),
            }
        } else {
            access_token
        };

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
    let url = format!("{}/api/room?user_id={}", BACKEND_URL, user_id);
    let access_token = get_data_from_keyring("access_token".to_string())?;
    let res = client
        .get(&url)
        .bearer_auth(&access_token)
        .send()
        .await
        .map_err(|e| e.to_string())?;

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
async fn fetch_room_messages(
    room_id: String,
    offset_id: String,
) -> Result<Vec<MessageResponse>, String> {
    let client = reqwest::Client::new();
    let url = format!("{}/api/fetch_room_message", BACKEND_URL);
    let access_token = get_data_from_keyring("access_token".to_string())?;

    let res = client
        .get(&url)
        .json(&serde_json::json!({
            "room_id": room_id,
            "offset_id": offset_id
        }))
        .bearer_auth(&access_token)
        .send()
        .await
        .map_err(|e| e.to_string())?;

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

#[tauri::command]
#[specta::specta]
async fn join_room(user_id: String, invite_code: String) -> Result<bool, String> {
    let client = reqwest::Client::new();
    let access_token = get_data_from_keyring("access_token".to_string())?;

    let url = format!("{}/api/join", BACKEND_URL);

    let res = client
        .post(&url)
        .bearer_auth(&access_token)
        .json(&serde_json::json!({"user_id": user_id, "invite_code": invite_code}))
        .send()
        .await
        .map_err(|e| e.to_string())?;

    if res.status().is_success() {
        Ok(true)
    } else {
        let error_body = res
            .text()
            .await
            .unwrap_or_else(|_| "Cannot read error body".to_string());
        println!("{}", error_body);
        Ok(false)
    }
}

#[tauri::command]
#[specta::specta]
async fn create_room(user_id: String, room_name: String) -> Result<bool, String> {
    let client = reqwest::Client::new();
    let access_token = get_data_from_keyring("access_token".to_string())?;

    let url = format!("{}/api/room", BACKEND_URL);

    let res = client
        .post(&url)
        .bearer_auth(&access_token)
        .json(&serde_json::json!({"user_id": user_id, "room_name": room_name}))
        .send()
        .await
        .map_err(|e| e.to_string())?;

    if res.status().is_success() {
        Ok(true)
    } else {
        let error_body = res
            .text()
            .await
            .unwrap_or_else(|_| "Cannot read error body".to_string());
        println!("{}", error_body);
        Ok(false)
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
        stop_tx: Arc::new(Mutex::new(None)),
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
        fetch_room_messages,
        register,
        checkLoginStatus,
        join_room,
        create_room
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
