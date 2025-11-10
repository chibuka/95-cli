use reqwest::blocking::{Client, multipart};
use serde::Deserialize;

const API_URL: &str = "https://api.ninefive.dev";
const FRONTEND_URL: &str = "https://ninefive.dev";

#[derive(Debug, Deserialize)]
#[allow(dead_code)]
pub struct SubmissionResponse {
    pub id: i64,
    #[serde(rename = "challengeId")]
    pub challenge_id: i64,
    #[serde(rename = "userId")]
    pub user_id: i64,
    #[serde(rename = "programmingLanguage")]
    pub programming_language: String,
    #[serde(rename = "executionResult")]
    pub execution_result: String,
    #[serde(rename = "targetStage")]
    pub target_stage: i32,
    #[serde(rename = "submittedAt")]
    pub submitted_at: String,
    #[serde(rename = "stageResultsJson")]
    pub stage_results_json: Option<String>,
    #[serde(rename = "executionOutput")]
    pub execution_output: Option<String>,
}

#[derive(Debug)]
pub struct SubmissionMetadata {
    pub id: i64,
    pub challenge_name: String,
    #[allow(dead_code)]
    pub challenge_slug: String,
    pub stage: i32,
    pub language: String,
    #[allow(dead_code)]
    pub url: String,
}

/// Submit code to the ninefive API and return full metadata
pub fn submit_code(submission_code: &str, zip_data: &[u8], verbose: bool) -> anyhow::Result<SubmissionMetadata> {
    let client = Client::new();

    let part = multipart::Part::bytes(zip_data.to_vec())
        .file_name("project.zip")
        .mime_str("application/zip")?;

    let form = multipart::Form::new()
        .text("submissionCode", submission_code.to_string())
        .part("project", part);

    let url = format!("{}/api/submissions/cli", API_URL);

    let response = client
        .post(&url)
        .multipart(form)
        .send()?;

    if !response.status().is_success() {
        let status = response.status();
        let error_text = response.text()?;
        return Err(anyhow::anyhow!("API error ({}): {}", status, error_text));
    }

    let response_text = response.text()?;
    let submission: SubmissionResponse = serde_json::from_str(&response_text)?;

    let debug_info = if verbose {
        Some((response_text.clone(), submission.id, submission.challenge_id, submission.programming_language.clone(), submission.target_stage))
    } else {
        None
    };

    let challenge_url_result = get_challenge_metadata(submission_code, verbose);

    let (challenge_name, challenge_slug, challenge_url, challenge_error) = match challenge_url_result {
        Ok((name, slug, url)) => (name, slug, url, None),
        Err(e) => {
            (
                "Unknown Challenge".to_string(),
                "unknown".to_string(),
                format!("{}/challenges", FRONTEND_URL),
                Some(e.to_string())
            )
        },
    };

    if verbose {
        if let Some((resp, id, chal_id, lang, stage)) = debug_info {
            eprintln!("\n=== DEBUG: SUBMISSION API RESPONSE ===");
            eprintln!("{}\n", resp);
            eprintln!("=== DEBUG: PARSED SUBMISSION DATA ===");
            eprintln!("ID: {}", id);
            eprintln!("Challenge ID: {}", chal_id);
            eprintln!("Language: {}", lang);
            eprintln!("Stage: {}", stage);
            if let Some(err) = challenge_error {
                eprintln!("\n=== DEBUG: CHALLENGE FETCH ERROR ===");
                eprintln!("{}\n", err);
            }
        }
    }

    Ok(SubmissionMetadata {
        id: submission.id,
        challenge_name,
        challenge_slug,
        stage: submission.target_stage,
        language: submission.programming_language,
        url: challenge_url,
    })
}

#[derive(Debug, Deserialize)]
struct JwtPayload {
    #[serde(rename = "challengeId")]
    challenge_id: i64,
    language: String,
}

#[derive(Debug, Deserialize)]
struct Challenge {
    slug: String,
    name: String,
}

/// Decode submission code and get full challenge metadata
fn get_challenge_metadata(submission_code: &str, verbose: bool) -> anyhow::Result<(String, String, String)> {
    let token = submission_code.strip_prefix("95-").unwrap_or(submission_code);

    // JWT format: header.payload.signature
    let parts: Vec<&str> = token.split('.').collect();
    if parts.len() != 3 {
        return Err(anyhow::anyhow!("Invalid submission code format"));
    }

    let payload_base64 = parts[1];
    let payload_bytes = base64::Engine::decode(
        &base64::engine::general_purpose::URL_SAFE_NO_PAD,
        payload_base64
    )?;
    let payload_str = String::from_utf8(payload_bytes)?;
    let payload: JwtPayload = serde_json::from_str(&payload_str)?;

    if verbose {
        eprintln!("\n=== DEBUG: JWT TOKEN ===");
        eprintln!("Payload: {}", payload_str);
        eprintln!("Challenge ID: {}", payload.challenge_id);
        eprintln!("Language: {}", payload.language);
    }

    let client = Client::new();
    let url = format!("{}/api/challenges/{}", API_URL, payload.challenge_id);

    if verbose {
        eprintln!("\n=== DEBUG: FETCHING CHALLENGE ===");
        eprintln!("GET {}", url);
    }

    let response = client.get(&url).send()?;
    let status = response.status();

    if verbose {
        eprintln!("Status: {}", status);
    }

    if !status.is_success() {
        let error_text = response.text()?;
        if verbose {
            eprintln!("Error: {}", error_text);
        }
        return Err(anyhow::anyhow!("API returned {}: {}", status, error_text));
    }

    let challenge_text = response.text()?;

    if verbose {
        eprintln!("Response: {}", challenge_text);
    }

    let challenge: Challenge = serde_json::from_str(&challenge_text)?;

    let language_lower = payload.language.to_lowercase();
    let challenge_url = format!(
        "{}/challenges/{}/{}",
        FRONTEND_URL,
        challenge.slug,
        language_lower
    );

    if verbose {
        eprintln!("\n=== DEBUG: CHALLENGE METADATA ===");
        eprintln!("Name: {}", challenge.name);
        eprintln!("Slug: {}", challenge.slug);
        eprintln!("URL: {}\n", challenge_url);
    }

    Ok((challenge.name, challenge.slug, challenge_url))
}
