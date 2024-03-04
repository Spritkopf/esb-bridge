use clap::Parser;
use esb_bridge::esb_bridge::Bridge;
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

    println!("Connecting to device {}", args.device);
    let bridge = Bridge::new(args.device, args.port).expect("Could not open Serial Port");

    let bridge_version = bridge::get_firmware_version().expect("Failed to read Firmware version");
    println!("esb-bridge firmware version: {bridge_version}");

    println!("Starting esb-bridge server...");
    println!("Listening to port {:?}", args.port);
    println!("Verbosity level: {:?}", args.verbose);


    
    
}