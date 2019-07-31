//ocelot creds k8s add -acct my_kewl_acct -name cluster_name -kubeconfig=/home/user/.kube/cluster-config.yaml

extern crate structopt;
use structopt::StructOpt;

use futures::Future;
use hyper::client::connect::{Destination, HttpConnector};
use tower_grpc::Request;
use tower_hyper::{client, util};
use tower_util::MakeService;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct AddOption {
    /// Account to add to. Defaults to auto-detect from current working directory
    #[structopt(name = "Account", long = "acct")]
    account: Option<String>,
    /// Kubernetes cluster name (logical)
    #[structopt(name = "Kubernetes cluster name", long)]
    cluster_name: Option<String>,

    /// Path to local repo. Defaults to current working directory
    #[structopt(long)]
    path: Option<String>,
    /// File path to Kubernetes config file
    #[structopt(name = "Kubernetes config (yaml)", short = "f", long = "file")]
    file_path: Option<String>,
}

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct DeleteOption {
    /// Account to add to. Defaults to auto-detect from current working directory
    #[structopt(name = "Account", long = "acct")]
    account: Option<String>,
}

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct ListOption {
    /// Account to add to. Defaults to auto-detect from current working directory
    #[structopt(name = "Account", long = "acct")]
    account: Option<String>,
}

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub enum ResourceAction {
    ///
    Add(AddOption),
    ///
    #[structopt(alias = "rm")]
    Delete(DeleteOption),
    ///
    #[structopt(alias = "ls")]
    List(ListOption),
}

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubOption {
    #[structopt(flatten)]
    action: ResourceAction,

    #[structopt(name = "Account", long = "acct")]
    account: Option<String>,
}

// Handle the command line control flow
pub fn subcommand_handler(args: SubOption) {
    let uri = ocelot_api::client_util::get_client_uri();
    let dst = Destination::try_from_uri(uri.clone()).unwrap();

    let connector = util::Connector::new(HttpConnector::new(4));
    let settings = client::Builder::new().http2_only(true).clone();
    let mut make_client = client::Connect::with_builder(connector, settings);

    match args.action {
        ResourceAction::Add(args) => {
            use git_meta::git_info;

            let identifier = args.cluster_name.clone().unwrap();
            let path_to_repo = ocelot_api::client_util::get_repo(args.path.clone());
            let file_contents = ocelot_api::client_util::read_file(args.file_path.clone());
            let git_info = git_info::get_git_info_from_path(&path_to_repo, &None, &None);

            let req = make_client
                .make_service(dst)
                .map_err(|e| panic!("connect error: {:?}", e))
                .and_then(move |conn| {
                    use ocelot_api::protobuf_api::legacyapi::client;

                    let conn = tower_request_modifier::Builder::new()
                        .set_origin(uri)
                        .build(conn)
                        .unwrap();

                    // Wait until the client is ready...
                    client::GuideOcelot::new(conn).ready()
                })
                .and_then(move |mut client| {
                    use ocelot_api::protobuf_api::legacyapi::K8sCreds;
                    use ocelot_api::protobuf_api::legacyapi::SubCredType;

                    let mut k8s_proto = K8sCreds::default();

                    k8s_proto.acct_name = git_info.account;
                    k8s_proto.sub_type = SubCredType::Kubeconf.into();
                    k8s_proto.identifier = identifier.to_string();
                    k8s_proto.k8s_contents = file_contents;

                    client.set_k8s_creds(Request::new(k8s_proto))
                })
                .and_then(|response| {
                    println!("RESPONSE = {:?}", response);
                    Ok(())
                })
                .map_err(|e| {
                    println!("ERR = {:?}", e);
                });

            tokio::run(req);
        }
        ResourceAction::Delete(_args) => {
            println!("There is no grpc endpoint for deleting kubernetes configs")
        }
        ResourceAction::List(_args) => {
            let req = make_client
                .make_service(dst)
                .map_err(|e| panic!("connect error: {:?}", e))
                .and_then(move |conn| {
                    use ocelot_api::protobuf_api::legacyapi::client;

                    let conn = tower_request_modifier::Builder::new()
                        .set_origin(uri)
                        .build(conn)
                        .unwrap();

                    // Wait until the client is ready...
                    client::GuideOcelot::new(conn).ready()
                })
                .and_then(move |mut client| client.get_generic_creds(Request::new(())))
                .and_then(|response| {
                    println!("RESPONSE = {:?}", response);
                    Ok(())
                })
                .map_err(|e| {
                    println!("ERR = {:?}", e);
                });

            tokio::run(req)
        }
    }
}
