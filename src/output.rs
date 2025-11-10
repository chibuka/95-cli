use console::style;
use indicatif::{ProgressBar, ProgressStyle};
use std::time::Duration;

#[allow(dead_code)]
pub fn info(msg: &str) {
    println!("  {} {}", style("→").cyan(), style(msg).white());
}

pub fn success(msg: &str) {
    println!("  {} {}", style("✓").green().bold(), style(msg).white());
}

pub fn error(msg: &str) {
    eprintln!("\n  {} {}\n", style("×").red().bold(), style(msg).white());
}

pub fn spinner(msg: &str) -> ProgressBar {
    let pb = ProgressBar::new_spinner();
    pb.set_style(
        ProgressStyle::default_spinner()
            .template("  {spinner:.cyan} {msg}")
            .unwrap()
            .tick_chars("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"),
    );
    pb.set_message(msg.to_string());
    pb.enable_steady_tick(Duration::from_millis(100));
    pb
}

#[allow(dead_code)]
pub fn step(number: usize, total: usize, msg: &str) {
    println!("  {} {}",
        style(format!("[{}/{}]", number, total)).dim(),
        style(msg).white()
    );
}
