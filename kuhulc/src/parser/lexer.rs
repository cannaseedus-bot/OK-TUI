#[derive(Clone, Debug)]
pub struct Token {
    pub lexeme: String,
}

pub fn lex(source: &str) -> anyhow::Result<Vec<Token>> {
    let tokens = source
        .lines()
        .map(str::trim)
        .filter(|line| !line.is_empty())
        .map(|lexeme| Token {
            lexeme: lexeme.to_string(),
        })
        .collect();

    Ok(tokens)
}
