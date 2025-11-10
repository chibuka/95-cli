use std::fs::File;
use std::io::{Read, Write};
use std::path::Path;
use walkdir::WalkDir;
use zip::write::SimpleFileOptions;
use zip::ZipWriter;

/// Create a ZIP archive from a directory
pub fn create_zip(dir: &Path, verbose: bool) -> anyhow::Result<Vec<u8>> {
    let mut buffer = Vec::new();
    {
        let mut zip = ZipWriter::new(std::io::Cursor::new(&mut buffer));
        let options = SimpleFileOptions::default()
            .compression_method(zip::CompressionMethod::Deflated)
            .unix_permissions(0o755);
        let walker = WalkDir::new(dir)
            .into_iter()
            .filter_entry(|e| !is_excluded(e.path()));

        for entry in walker {
            let entry = entry?;
            let path = entry.path();
            let name = path.strip_prefix(dir)?;

            if path.is_file() {
                if verbose {
                    println!("  Adding: {}", name.display());
                }

                zip.start_file(name.to_string_lossy().to_string(), options)?;

                let mut file = File::open(path)?;
                let mut file_buffer = Vec::new();
                file.read_to_end(&mut file_buffer)?;

                zip.write_all(&file_buffer)?;
            } else if !name.as_os_str().is_empty() && path.is_dir() {
                if verbose {
                    println!("  Adding directory: {}", name.display());
                }
                zip.add_directory(name.to_string_lossy().to_string(), options)?;
            }
        }

        zip.finish()?;
    }

    Ok(buffer)
}

/// Check if a path should be excluded from the ZIP
fn is_excluded(path: &Path) -> bool {
    let excluded_names = [
        ".git",
        ".gitignore",
        "node_modules",
        "target",
        ".DS_Store",
        "__pycache__",
        ".pytest_cache",
        ".vscode",
        ".idea",
        "*.swp",
        "*.swo",
        ".ninefiveignore",
    ];

    path.file_name()
        .and_then(|name| name.to_str())
        .map(|name| excluded_names.iter().any(|&excl| name == excl || name.starts_with(excl)))
        .unwrap_or(false)
}
