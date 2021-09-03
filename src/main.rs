use roperator::{
    prelude::{Error, Handler, K8sType, OperatorConfig, SyncRequest, SyncResponse},
    runner,
    serde_json::json,
};

const OPERATOR_NAME: &str = "wasmcloud-k8s-operator";

static RESOURCE_TYPE: &K8sType = &K8sType {
    api_version: "core.oam.dev/v1beta1",
    kind: "Application",
    plural_kind: "applications",
};

struct MyHandler;
impl Handler for MyHandler {
    fn sync(&self, _request: &SyncRequest) -> Result<SyncResponse, Error> {
        // TODO: place a message to topic

        let status = json!({
            "message": "all good mate!",
            "phase": "Running",
        });

        Ok(SyncResponse {
            status,
            children: Vec::new(),
            resync: None,
        })
    }
}

fn main() {
    let operator_config = OperatorConfig::new(OPERATOR_NAME, RESOURCE_TYPE);

    let err = runner::run_operator(operator_config, MyHandler);

    eprintln!("Error running operator: {}", err);
    std::process::exit(1);
}
