use roperator::{
    config::ClientConfig,
    prelude::{Error, Handler, K8sType, OperatorConfig, SyncRequest, SyncResponse},
    runner,
    serde_json::json,
};

const OPERATOR_NAME: &str = "wasmcloud-k8s-operator";

static RESOURCE_TYPE: &K8sType = &K8sType {
    api_version: "wasmcloud.com/v1beta1",
    kind: "WasmCloudApplication",
    plural_kind: "wasmcloudapplications",
};

struct MyHandler;
impl Handler for MyHandler {
    fn sync(&self, request: &SyncRequest) -> Result<SyncResponse, Error> {
        // TODO: place a message to topic

        // dbg!(request);

        Ok(SyncResponse {
            status: json!({
                "message": "magical",
                "phase": "Running",
            }),
            children: Vec::new(),
            resync: None,
        })
    }
}

fn main() {
    env_logger::init();
    let operator_config = OperatorConfig::new(OPERATOR_NAME, RESOURCE_TYPE);

    let client_config = ClientConfig::from_kubeconfig(OPERATOR_NAME.to_string())
        .expect("failed to resolve cluster data from kubeconfig");

    let err = runner::run_operator_with_client_config(operator_config, client_config, MyHandler);

    eprintln!("Error running operator: {}", err);
    std::process::exit(1);
}
