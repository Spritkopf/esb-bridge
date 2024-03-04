

use std::time::Duration;

use clap::Parser;
use serialport;
use esb_bridge::bridge::Bridge;

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
    let mut my_bridge = Bridge::new(args.device, args.port);

    
    
    println!("Starting esb-bridge server...");
    println!("Listening to port {:?}", args.port);
    println!("Verbosity level: {:?}", args.verbose);
    
    let bridge_version = my_bridge.get_firmware_version().expect("Failed to read Firmware version");
    println!("esb-bridge firmware version: {bridge_version}");
    let test = my_bridge.transfer().unwrap();
    
}