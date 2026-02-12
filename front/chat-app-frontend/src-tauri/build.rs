fn main() {
    // Load .env file at build time
    if let Err(e) = dotenvy::dotenv() {
        println!("cargo:warning=Failed to load .env file: {}", e);
    }
    
    // Read the environment variables and pass them to rustc
    if let Ok(val) = std::env::var("BACKEND_REGISTER_URL") {
        println!("cargo:rustc-env=BACKEND_REGISTER_URL={}", val);
    }
    if let Ok(val) = std::env::var("REDIRECT_URL") {
        println!("cargo:rustc-env=REDIRECT_URL={}", val);
    }
    if let Ok(val) = std::env::var("TOKEN_URL") {
        println!("cargo:rustc-env=TOKEN_URL={}", val);
    }
    if let Ok(val) = std::env::var("CLIENT_ID") {
        println!("cargo:rustc-env=CLIENT_ID={}", val);
    }
    
    // Rerun if these change
    // println!("cargo:rerun-if-env-changed=BACKEND_REGISTER_URL");
    // println!("cargo:rerun-if-env-changed=REDIRECT_URL");
    // println!("cargo:rerun-if-env-changed=TOKEN_URL");
    // println!("cargo:rerun-if-env-changed=CLIENT_ID");
    // println!("cargo:rerun-if-changed=.env");
    tauri_build::build();
}
