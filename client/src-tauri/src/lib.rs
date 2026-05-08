use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize, Default)]
pub struct AppInfo {
    pub name: String,
    pub version: String,
    pub commit: Option<String>,
}

#[tauri::command]
fn app_info() -> AppInfo {
    AppInfo {
        name: env!("CARGO_PKG_NAME").to_string(),
        version: env!("CARGO_PKG_VERSION").to_string(),
        commit: option_env!("ROUTEFLOW_COMMIT").map(str::to_string),
    }
}

#[derive(Debug, Serialize)]
pub struct Echo {
    pub message: String,
}

#[tauri::command]
fn ping(message: Option<String>) -> Echo {
    Echo {
        message: message.unwrap_or_else(|| "pong".to_string()),
    }
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_os::init())
        .plugin(tauri_plugin_dialog::init())
        .plugin(tauri_plugin_fs::init())
        .plugin(tauri_plugin_notification::init())
        .invoke_handler(tauri::generate_handler![app_info, ping])
        .run(tauri::generate_context!())
        .expect("error while running RouteFlow");
}
