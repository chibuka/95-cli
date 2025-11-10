mod api;
mod zip_util;
mod output;
mod ui;

use clap::{Parser, Subcommand};
use std::path::PathBuf;

#[derive(Parser)]
#[command(name = "95")]
#[command(about = "CLI for ninefive code challenge platform", long_about = None)]
#[command(version)]
struct Cli {
    #[command(subcommand)]
    command: Option<Commands>,
}

#[derive(Subcommand)]
enum Commands {
    /// Submit code to ninefive platform
    Submit {
        /// Submission code from the web interface (optional, will prompt if not provided)
        code: Option<String>,
        /// Directory to submit (defaults to current directory)
        #[arg(short, long, default_value = ".")]
        path: PathBuf,
        /// Show verbose output
        #[arg(short, long)]
        verbose: bool,
    },
}

fn main() {
    ui::print_banner();
    ui::show_current_directory();

    let cli = Cli::parse();

    match cli.command {
        Some(Commands::Submit { code, path, verbose }) => {
            // Get submission code from argument or prompt
            let submission_code = match code {
                Some(c) => c,
                None => match ui::prompt_submission_code() {
                    Ok(c) => c,
                    Err(e) => {
                        output::error(&format!("failed to read input: {}", e));
                        std::process::exit(1);
                    }
                },
            };

            if let Err(e) = handle_submit(submission_code, path, verbose) {
                output::error(&e.to_string());
                std::process::exit(1);
            }
        }
        None => {
            let submission_code = match ui::prompt_submission_code() {
                Ok(c) => c,
                Err(e) => {
                    output::error(&format!("failed to read input: {}", e));
                    std::process::exit(1);
                }
            };

            if let Err(e) = handle_submit(submission_code, PathBuf::from("."), false) {
                output::error(&e.to_string());
                std::process::exit(1);
            }
        }
    }
}

fn handle_submit(code: String, path: PathBuf, verbose: bool) -> anyhow::Result<()> {
    println!();

    let spinner = output::spinner("creating archive");
    let zip_data = match zip_util::create_zip(&path, verbose) {
        Ok(data) => data,
        Err(e) => {
            spinner.finish_and_clear();
            return Err(anyhow::anyhow!("failed to create archive: {}", format_user_error(&e)));
        }
    };
    spinner.finish_and_clear();
    output::success(&format!("archive created ({} bytes)", zip_data.len()));

    let spinner = output::spinner("uploading submission");
    let metadata = match api::submit_code(&code, &zip_data, verbose) {
        Ok(data) => data,
        Err(e) => {
            spinner.finish_and_clear();
            return Err(anyhow::anyhow!("{}", format_user_error(&e)));
        }
    };
    spinner.finish_and_clear();
    output::success(&format!("submission received (#{})", metadata.id));

    let details = ui::SubmissionDetails {
        challenge: metadata.challenge_name.clone(),
        stage: format!("{}", metadata.stage),
        language: metadata.language.clone(),
        id: format!("#{}", metadata.id),
    };
    ui::print_details_table(details);

    ui::success_box("submission is being processed",);

    Ok(())
}

fn format_user_error(err: &anyhow::Error) -> String {
    let err_str = err.to_string();

    if err_str.contains("401") || err_str.contains("Unauthorized") || err_str.contains("Invalid or expired") {
        return "invalid or expired submission code".to_string();
    }

    if err_str.contains("404") || err_str.contains("Not Found") {
        return "challenge not found".to_string();
    }

    if err_str.contains("500") || err_str.contains("Internal Server") {
        return "server error, please try again later".to_string();
    }

    if err_str.contains("Connection refused") || err_str.contains("connection") {
        return "unable to connect to ninefive.dev - please check your internet connection".to_string();
    }

    if err_str.contains("timeout") {
        return "request timed out, please try again".to_string();
    }

    if err_str.contains("No such file") || err_str.contains("not found") {
        return "directory not found".to_string();
    }

    if err_str.contains("Permission denied") {
        return "permission denied".to_string();
    }

    "something went wrong, please try again".to_string()
}
