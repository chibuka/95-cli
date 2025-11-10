use console::style;
use dialoguer::Input;

const VERSION: &str = env!("CARGO_PKG_VERSION");

pub fn print_banner() {
    let amber = console::Color::Color256(221); 

    println!();
    println!("{}", style(" █████╗ ███████╗").fg(amber).bold());
    println!("{}", style("██╔══██╗██╔════╝").fg(amber).bold());
    println!("{}", style("╚██████║███████╗").fg(amber).bold());
    println!("{}", style(" ╚═══██║╚════██║").fg(amber).bold());
    println!("{}", style(" █████╔╝███████║").fg(amber).bold());
    println!("{}", style(" ╚════╝ ╚══════╝").fg(amber).bold());
    println!("  {} {}", style("ninefive").bold(), style(format!("v{}", VERSION)).dim());
    println!("  {}", style("submit code challenges without manual zipping").dim());
    println!();
}

/// Show current working directory
pub fn show_current_directory() {
    let cwd = std::env::current_dir()
        .map(|p| p.display().to_string())
        .unwrap_or_else(|_| ".".to_string());

    println!("{}", style(cwd).dim());
    println!();
}

/// Prompt for submission code
pub fn prompt_submission_code() -> anyhow::Result<String> {
    use dialoguer::theme::ColorfulTheme;

    let theme = ColorfulTheme::default();

    let code: String = Input::with_theme(&theme)
        .with_prompt(format!("{} submission code", style(">").cyan().bold()))
        .validate_with(|input: &String| -> Result<(), &str> {
            let input = input.trim();
            if input.is_empty() {
                return Err("submission code cannot be empty");
            }

            let token = input.strip_prefix("95-").unwrap_or(input);
            let parts: Vec<&str> = token.split('.').collect();
            if parts.len() != 3 {
                return Err("invalid submission code format (expected JWT token)");
            }

            Ok(())
        })
        .interact_text()?;

    Ok(code.trim().to_string())
}

pub struct SubmissionDetails {
    pub challenge: String,
    pub stage: String,
    pub language: String,
    pub id: String,
}

/// Print submission details in a grid table
pub fn print_details_table(details: SubmissionDetails) {
    println!();
    println!("{}", style("submission details").bold().white());
    println!();

    let col_width = 12;

    println!("  {}", style("┌─────────────┬─────────────────────────────────┐").dim());

    println!("  {} {:<width$} {} {:<29} {}",
        style("│").dim(),
        style("challenge").dim(),
        style("│").dim(),
        details.challenge,
        style("│").dim(),
        width = col_width
    );

    println!("  {}", style("├─────────────┼─────────────────────────────────┤").dim());

    println!("  {} {:<width$} {} {:<29} {}",
        style("│").dim(),
        style("stage").dim(),
        style("│").dim(),
        details.stage,
        style("│").dim(),
        width = col_width
    );

    println!("  {}", style("├─────────────┼─────────────────────────────────┤").dim());

    println!("  {} {:<width$} {} {:<29} {}",
        style("│").dim(),
        style("language").dim(),
        style("│").dim(),
        details.language,
        style("│").dim(),
        width = col_width
    );

    println!("  {}", style("├─────────────┼─────────────────────────────────┤").dim());

    println!("  {} {:<width$} {} {:<29} {}",
        style("│").dim(),
        style("id").dim(),
        style("│").dim(),
        details.id,
        style("│").dim(),
        width = col_width
    );

    println!("  {}", style("└─────────────┴─────────────────────────────────┘").dim());
}


pub fn success_box(title: &str) {
    println!();
    println!("{}", style(title).green().bold());
    println!();
    println!("  {}", style("view progress via your browser.").dim());
    println!();
}

