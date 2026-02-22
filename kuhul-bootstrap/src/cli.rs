#[derive(Copy, Clone, Debug)]
pub enum Target {
    C,
    Gpu,
    Scxq2,
}

#[derive(Debug)]
pub struct Cli {
    pub input: String,
    pub target: Target,
    pub output: String,
    pub deterministic: bool,
}

impl Cli {
    pub fn parse() -> Result<Self, String> {
        let mut args = std::env::args().skip(1);
        let input = args.next().ok_or("missing input path")?;

        let mut target = Target::C;
        let mut output = None;
        let mut deterministic = false;

        while let Some(arg) = args.next() {
            match arg.as_str() {
                "--target" => {
                    let val = args.next().ok_or("missing value for --target")?;
                    target = match val.as_str() {
                        "c" => Target::C,
                        "gpu" => Target::Gpu,
                        "scxq2" => Target::Scxq2,
                        _ => return Err(format!("unsupported target: {val}")),
                    };
                }
                "--output" => output = args.next(),
                "--deterministic" => deterministic = true,
                _ => return Err(format!("unknown argument: {arg}")),
            }
        }

        let output = output.ok_or("missing --output path")?;
        Ok(Self {
            input,
            target,
            output,
            deterministic,
        })
    }
}
