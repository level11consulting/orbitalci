use color_eyre::eyre::Result;
use structopt::StructOpt;

use crate::subcommand::GlobalOption;

/// Start an Orb server
pub mod start;

/// Polling functionality
pub mod poll;

/// Subcommands for `orb server`
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub enum ServerType {
    Start(start::SubcommandOption),
}

/// Subcommand router for `orb server`
pub async fn subcommand_handler(
    global_option: GlobalOption,
    server_subcommand: ServerType,
) -> Result<()> {
    match server_subcommand {
        ServerType::Start(sub_option) => start::subcommand_handler(global_option, sub_option).await,
    }
}
