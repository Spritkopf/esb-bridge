

use std::process::exit;

use env_logger;
use log;
use clap::Parser;
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
}

fn main() {
    let args = Args::parse();
    let mut my_bridge: Bridge;

    env_logger::init();

    log::info!("Connecting to device {}", args.device);
    match Bridge::new(args.device, args.port) {
        Ok(bridge) => my_bridge = bridge,
        Err(msg) => {
            log::error!("Error opening connection to Bridge device: {}", msg);
            exit(-1);
        }
    }


    let bridge_version = my_bridge.get_firmware_version().expect("Failed to read Firmware version");
    log::info!("esb-bridge firmware version: {bridge_version}");

    log::info!("Starting esb-bridge server on port {}...", args.port);
    
}