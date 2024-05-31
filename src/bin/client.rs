
use std::process::exit;

use clap::{Parser, Subcommand};
use esb_bridge::bridge::Bridge;


#[derive(Parser)]
#[clap(name = "esbbridge-client", version)]
pub struct App {
    #[clap(value_parser, required = true)]
    /// Serial port for esb-bridge device (e.g. /dev/ttyACM0)
    device: String,

    #[clap(subcommand)]
    subcommand: Subcommands,
}

#[derive(Subcommand)]
enum Subcommands {
    Msg {
        #[clap(short, long)]
        // target ESB address in format XX:XX:XX:XX:XX
        target: String,
    }
}
fn main() {
    let args = App::parse();
    let mut my_bridge: Bridge;

    println!("Connecting to device {}", args.device);
    match Bridge::new(args.device) {
        Ok(bridge) => my_bridge = bridge,
        Err(msg) => {
            println!("Error opening connection to Bridge device: {}", msg);
            exit(-1);
        }
    }

    
    let bridge_version = my_bridge.get_firmware_version().expect("Failed to read Firmware version");
    println!("esb-bridge firmware version: {bridge_version}");
}
