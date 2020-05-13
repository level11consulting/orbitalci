use structopt::StructOpt;

use crate::{repo::SubcommandOption, GlobalOption};

use orbital_headers::code::{code_service_client::CodeServiceClient, GitRepoUpdateRequest};
use orbital_services::ORB_DEFAULT_URI;
use tonic::Request;

use git_meta::git_info;
use log::debug;
use std::path::PathBuf;

use anyhow::Result;
use orbital_database::postgres::repo::Repo;
use prettytable::{cell, format, row, Table};

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {
    /// Repo path
    #[structopt(parse(from_os_str), env = "PWD")]
    path: PathBuf,

    /// Name of Orbital org
    #[structopt(long, env = "ORB_DEFAULT_ORG")]
    org: Option<String>,
}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    action_option: ActionOption,
) -> Result<()> {
    let repo_info =
        match git_info::get_git_info_from_path(&action_option.path.as_path(), &None, &None) {
            Ok(info) => info,
            Err(_e) => panic!("Unable to parse path for git repo info"),
        };

    let request = Request::new(GitRepoUpdateRequest {
        org: action_option.org.unwrap_or_default(),
        git_provider: repo_info.git_url.host.clone().unwrap(),
        name: repo_info.git_url.name.clone(),
        //user: repo_info,
        uri: repo_info.git_url.trim_auth().to_string(),
        //secret_type: ,
        //build: ,
        //notify: ,
        //auth_data: ,
        ..Default::default()
    });

    debug!("Request for git repo update: {:?}", &request);

    let mut client = CodeServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    let response = client.git_repo_update(request).await;

    match response {
        Err(_e) => {
            eprintln!("Repo not found");
            Ok(())
        }
        Ok(o) => {
            let repo_proto = o.into_inner();

            debug!("RESPONSE = {:?}", &repo_proto);

            // By default, format the response into a table
            let mut table = Table::new();
            table.set_format(*format::consts::FORMAT_NO_BORDER_LINE_SEPARATOR);

            // Print the header row
            table.set_titles(row![
                bc =>
                "Org Name",
                "Repo Name",
                "Uri",
                "Secret Name",
                "Build Enabled",
                "Notify Enabled",
                "Next build index"
            ]);

            let repo = Repo::from(repo_proto.clone());

            table.add_row(row![
                repo.org_id,
                repo.name,
                repo.uri,
                repo_proto.auth_data,
                &format!("{:?}", repo.build_active_state),
                &format!("{:?}", repo.notify_active_state),
                repo.next_build_index
            ]);

            // Print the table to stdout
            table.printstd();
            Ok(())
        }
    }
}
