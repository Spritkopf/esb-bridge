use clap::Parser;
#[derive(Parser, Debug)]
#[clap(author, version, about, long_about = None)]
struct Args {
    /// TCP port for incoming client connections
    #[clap(short, long, parse(try_from_str))]
    port: u32,

    /// Serial device for esb-bridge device (e.g. /dev/ttyACM0)
    #[clap(short, long)]
    device: String,

    /// Additional output (up to three levels)
    #[clap(short, long, parse(from_occurrences))]
    verbose: usize,
}

fn main() {
    let args = Args::parse();

    println!("Starting esb-bridge server...");
    println!("Connecting to device {}", args.device);
    println!("Listening to port {:?}", args.port);
    println!("Verbosity level: {:?}", args.verbose);
}