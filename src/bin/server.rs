use clap::Parser;
use esb_bridge::server::usb_protocol::Message;

#[derive(Parser, Debug)]
#[clap(author, version, about, long_about = None)]
struct Args {
    /// Name of the person to greet
    #[clap(short, long)]
    name: String,

    /// Number of times to greet
    #[clap(short, long, default_value_t = 1)]
    count: u8,
}

fn main() {
    // let msg = Message::new(12, vec![1,2,3]).unwrap();
    // println!("Hello Message {:?}", msg.payload);

    let args = Args::parse();

    for _ in 0..args.count {
        println!("Hello {}!", args.name)
    }
}